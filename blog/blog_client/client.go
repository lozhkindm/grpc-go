package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect: %v\n", err)
	}
	defer func(cc *grpc.ClientConn) {
		if err := cc.Close(); err != nil {
			log.Fatalf("Cannot close a client connection: %v\n", err)
		}
	}(cc)

	cl := blogpb.NewBlogServiceClient(cc)
	blogId := createBlog(cl)
	readBlog(cl, blogId)
	updateBlog(cl, blogId)
	deleteBlog(cl, blogId)
	readAllBlogs(cl)
}

func createBlog(cl blogpb.BlogServiceClient) string {
	req := &blogpb.CreateBlogRequest{Blog: &blogpb.Blog{
		AuthorId: "Vasya",
		Title:    "How to be a human?",
		Content:  "I do not know...",
	}}
	res, err := cl.CreateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while creating a blog: %v\n", err)
	}
	log.Printf("Response from CreateBlog: %v\n", res)

	return res.GetBlog().GetId()
}

func readBlog(cl blogpb.BlogServiceClient, blogId string) {
	res, err := cl.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: blogId})
	if err != nil {
		log.Fatalf("Error while reading a blog: %v\n", err)
	}
	log.Printf("Response from ReadBlog: %v\n", res)
}

func updateBlog(cl blogpb.BlogServiceClient, blogId string) {
	req := &blogpb.UpdateBlogRequest{Blog: &blogpb.Blog{
		Id:       blogId,
		AuthorId: "Olesha",
		Title:    "How to be a robot?",
		Content:  "I know...",
	}}
	res, err := cl.UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while updating a blog: %v\n", err)
	}
	log.Printf("Response from UpdateBlog: %v\n", res)
}

func deleteBlog(cl blogpb.BlogServiceClient, blogId string) {
	res, err := cl.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogId})
	if err != nil {
		log.Fatalf("Error while deleting a blog: %v\n", err)
	}
	log.Printf("Response from DeleteBlog: %v\n", res)
}

func readAllBlogs(cl blogpb.BlogServiceClient) {
	stream, err := cl.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Error while reading the blogs: %v\n", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v\n", err)
		}
		log.Printf("Response from ListBlog: %v\n", res)
	}
}
