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
