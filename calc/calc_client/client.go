package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/calc/calcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer func(cc *grpc.ClientConn) {
		err := cc.Close()
		if err != nil {
			log.Fatalf("Cannot close a client connection: %v", err)
		}
	}(cc)

	cl := calcpb.NewCalcServiceClient(cc)
	makeSumCall(cl)
	makePrimeCall(cl)
	makeAverageCall(cl)
	makeMaximumCall(cl)
	makeRootCall(cl)
}

func makeSumCall(cl calcpb.CalcServiceClient) {
	req := &calcpb.CalcRequest{
		Number1: 3,
		Number2: 10,
	}

	res, err := cl.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Sum RPC: %v", err)
	}
	log.Printf("Response from Sum: %v", res.GetResult())
}

func makePrimeCall(cl calcpb.CalcServiceClient) {
	req := &calcpb.PrimeRequest{Number: 120}

	stream, err := cl.Prime(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Prime RPC: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		log.Printf("Response from Prime: %v", res.GetResult())
	}
}

func makeAverageCall(cl calcpb.CalcServiceClient) {
	numbers := []int32{1, 2, 3, 4}

	stream, err := cl.Average(context.Background())
	if err != nil {
		log.Fatalf("Error while calling Average RPC: %v", err)
	}

	for _, n := range numbers {
		if err := stream.Send(&calcpb.AverageRequest{Number: n}); err != nil {
			log.Fatalf("Error while sending a request to the stream: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from Average: %v", err)
	}
	log.Printf("Response from Average: %v", res.GetResult())
}

func makeMaximumCall(cl calcpb.CalcServiceClient) {
	stream, err := cl.Maximum(context.Background())
	if err != nil {
		log.Fatalf("Error while calling Maximum RPC: %v", err)
	}

	numbers := []int32{1, 5, 3, 6, 2, 20}
	wc := make(chan struct{})

	go func() {
		for _, number := range numbers {
			if err := stream.Send(&calcpb.MaximumRequest{Number: number}); err != nil {
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
			log.Printf("Response from Maximum: %v", res.GetResult())
		}
		close(wc)
	}()

	<-wc
}

func makeRootCall(cl calcpb.CalcServiceClient) {
	numbers := []int32{10, -10}
	for _, num := range numbers {
		rootForNumber(cl, num)
	}
}

func rootForNumber(cl calcpb.CalcServiceClient, num int32) {
	res, err := cl.Root(context.Background(), &calcpb.RootRequest{Number: num})
	if err != nil {
		if err, ok := status.FromError(err); ok {
			log.Fatalln(err)
		} else {
			log.Fatalf("Error while calling Root RPC: %v\n", err)
		}
	}
	log.Printf("Response from Root: %v\n", res.GetResult())
}
