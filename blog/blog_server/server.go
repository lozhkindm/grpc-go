package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"os/signal"
)

var collection *mongo.Collection

type BlogItem struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorId string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

type Server struct{}

func (Server) CreateBlog(_ context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()

	item := BlogItem{
		AuthorId: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	one, err := collection.InsertOne(context.Background(), item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal Server Error: %v", err)
	}

	oid, ok := one.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Error(codes.Internal, "Cannot cast to ObjectID")
	}

	res := &blogpb.CreateBlogResponse{Blog: &blogpb.Blog{
		Id:       oid.Hex(),
		AuthorId: blog.AuthorId,
		Title:    blog.Title,
		Content:  blog.Content,
	}}

	return res, nil
}

func (Server) ReadBlog(_ context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot cast to ObjectID: %v", err)
	}

	item := &BlogItem{}
	one := collection.FindOne(context.Background(), primitive.M{"_id": oid})
	if err := one.Decode(item); err != nil {
		return nil, status.Errorf(codes.NotFound, "Blog not found: %v", err)
	}

	return &blogpb.ReadBlogResponse{Blog: makeBlogPb(item)}, nil
}

func (Server) UpdateBlog(_ context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot cast to ObjectID: %v", err)
	}

	item := &BlogItem{}
	filter := primitive.M{"_id": oid}
	one := collection.FindOne(context.Background(), filter)
	if err := one.Decode(item); err != nil {
		return nil, status.Errorf(codes.NotFound, "Blog not found: %v", err)
	}

	item.AuthorId = blog.GetAuthorId()
	item.Title = blog.GetTitle()
	item.Content = blog.GetContent()

	_, err = collection.ReplaceOne(context.Background(), filter, item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot update the blog: %v", err)
	}

	return &blogpb.UpdateBlogResponse{Blog: makeBlogPb(item)}, nil
}

func (Server) DeleteBlog(_ context.Context, req *blogpb.DeleteBlogRequest) (*emptypb.Empty, error) {
	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot cast to ObjectID: %v", err)
	}

	filter := primitive.M{"_id": oid}
	one, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot delete the blog: %v", err)
	}
	if one.DeletedCount == 0 {
		return nil, status.Error(codes.NotFound, "Blog not found")
	}

	return &emptypb.Empty{}, nil
}

func (Server) ListBlog(_ *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	cursor, err := collection.Find(context.Background(), primitive.M{})
	if err != nil {
		return status.Errorf(codes.Internal, "Cannot find blogs: %v", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			log.Fatalf("Error while closing a cursor: %v", err)
		}
	}(cursor, context.Background())

	for cursor.Next(context.Background()) {
		item := &BlogItem{}
		if err := cursor.Decode(item); err != nil {
			return status.Errorf(codes.Internal, "Cannot decode a blog: %v", err)
		}
		res := &blogpb.ListBlogResponse{Blog: makeBlogPb(item)}
		if err := stream.Send(res); err != nil {
			return status.Errorf(codes.Internal, "Cannot send a blog: %v", err)
		}
	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, "Unexpected error: %v", err)
	}

	return nil
}

func makeBlogPb(item *BlogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       item.Id.Hex(),
		AuthorId: item.AuthorId,
		Title:    item.Title,
		Content:  item.Content,
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to mongo: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Failed to disconnect from mongo: %v", err)
		}
	}()

	collection = client.Database("db_grpc").Collection("blogs")

	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(server, &Server{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

	server.Stop()
	if err := listener.Close(); err != nil {
		log.Fatalf("Failed to close the listener: %v", err)
	}
}
