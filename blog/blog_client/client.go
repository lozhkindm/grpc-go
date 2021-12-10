package main

import (
	"context"
	"fmt"
	"github.com/lozhkindm/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
	//createBlog(cl)
	//readBlog(cl)
	//updateBlog(cl)
	deleteBlog(cl)
}

func readBlog(cl blogpb.BlogServiceClient) {
	res, err := cl.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "61b251c9381d0c20a417fffc"})
	if err != nil {
		log.Fatalf("Error while reading a blog: %v", err)
	}
	fmt.Printf("Response from ReadBlog: %v", res)
}

func createBlog(cl blogpb.BlogServiceClient) {
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

func updateBlog(cl blogpb.BlogServiceClient) {
	req := &blogpb.UpdateBlogRequest{Blog: &blogpb.Blog{
		Id:       "61b251c9381d0c20a417fffc",
		AuthorId: "Olesha",
		Title:    "How to be a robot?",
		Content:  "I know...",
	}}
	res, err := cl.UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while updating a blog: %v", err)
	}
	log.Printf("Response from UpdateBlog: %v", res)
}

func deleteBlog(cl blogpb.BlogServiceClient) {
	res, err := cl.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: "61b251c9381d0c20a417fffc"})
	if err != nil {
		log.Fatalf("Error while deleting a blog: %v", err)
	}
	log.Printf("Response from DeleteBlog: %v", res)
}
