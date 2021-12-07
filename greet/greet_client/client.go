package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
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
	makeGreetEveryoneCall(cl)
	makeGreetWithDeadlineCall(cl, time.Millisecond*5000)
	makeGreetWithDeadlineCall(cl, time.Millisecond*1000)
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

func makeGreetEveryoneCall(cl greetpb.GreetServiceClient) {
	stream, err := cl.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while calling GreetEveryone RPC: %v", err)
	}

	names := []string{"Ivan", "Jorik", "Vasya", "Olesha"}
	wc := make(chan struct{})

	go func() {
		for _, name := range names {
			req := &greetpb.GreetEveryoneRequest{
				Greeting: &greetpb.Greeting{FirstName: name},
			}
			if err := stream.Send(req); err != nil {
				log.Fatalf("Error while sending a request to the stream: %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("Error while closing the stream: %v", err)
		}
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while reading the stream: %v", err)
			}
			log.Printf("Response from GreetEveryone: %v", res.GetResult())
		}
		close(wc)
	}()

	<-wc
}

func makeGreetWithDeadlineCall(cl greetpb.GreetServiceClient, timeout time.Duration) {
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{FirstName: "Vasya", LastName: "Pupkin"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := cl.GreetWithDeadline(ctx, req)
	if err != nil {
		if err, ok := status.FromError(err); ok {
			if err.Code() == codes.DeadlineExceeded {
				log.Fatalf("Deadline exceeded: %v", err)
			} else {
				log.Fatalf("Unexpected error: %v", err)
			}
		} else {
			log.Fatalf("Error while calling GreetWithDeadline RPC: %v", err)
		}
	}
	log.Printf("Response from GreetWithDeadline: %v", res.GetResult())
}
