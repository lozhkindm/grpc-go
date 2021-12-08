package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
)

var collection *mongo.Collection

type Server struct{}

type BlogItem struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorId string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
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
