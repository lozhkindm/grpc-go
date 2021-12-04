package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer func(cc *grpc.ClientConn) {
		if err := cc.Close(); err != nil {
			log.Fatalf("Cannot close a client connection: %v", err)
		}
	}(cc)

	cl := greetpb.NewGreetServiceClient(cc)
	makeGreetCall(cl)
	makeGreetManyTimesCall(cl)
	makeLongGreetCall(cl)
}

func makeGreetCall(cl greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Nikolay",
			LastName:  "Valuev",
		},
	}

	res, err := cl.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.GetResult())
}

func makeGreetManyTimesCall(cl greetpb.GreetServiceClient) {
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Engeniy",
			LastName:  "Onegin",
		},
	}

	stream, err := cl.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", res.GetResult())
	}
}

func makeLongGreetCall(cl greetpb.GreetServiceClient) {
	names := []string{"Nikolay", "Evgeniy", "Boris", "Timur", "Alex"}

	stream, err := cl.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet RPC: %v", err)
	}

	for _, name := range names {
		req := &greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{FirstName: name},
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("Error while sending a request to the stream: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet: %v", err)
	}
	log.Printf("Response from LongGreet: %v", res.GetResult())
}
