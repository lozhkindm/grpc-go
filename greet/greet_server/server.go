package main

import (
	"context"
	"fmt"
	"github.com/lozhkindm/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"strings"
)

type server struct{}

func (server) Greet(_ context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	result := fmt.Sprintf("Hello, %s", req.GetGreeting().GetFirstName())
	return &greetpb.GreetResponse{Result: result}, nil
}

func (server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	for i := 0; i < 10; i++ {
		res := &greetpb.GreetManyTimesResponse{
			Result: fmt.Sprintf("Hello %s, number %d", req.GetGreeting().GetFirstName(), i),
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
	return nil
}

func (server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	var names []string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := &greetpb.LongGreetResponse{
				Result: fmt.Sprintf("Hello: %s", strings.Join(names, ", ")),
			}
			return stream.SendAndClose(res)
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		names = append(names, req.GetGreeting().GetFirstName())
	}
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(srv, &server{})

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
