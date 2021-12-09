package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
	"log"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	defer func(cc *grpc.ClientConn) {
		if err := cc.Close(); err != nil {
			log.Fatalf("Cannot close a client connection: %v", err)
		}
	}(cc)

	cl := blogpb.NewBlogServiceClient(cc)
	req := &blogpb.CreateBlogRequest{Blog: &blogpb.Blog{
		AuthorId: "Vasya",
		Title:    "How to be a human?",
		Content:  "I do not know...",
	}}
	res, err := cl.CreateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while creating a blog: %v", err)
	}

	log.Printf("Response from CreateBlog: %v", res)
}
