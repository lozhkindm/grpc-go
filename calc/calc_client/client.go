package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/calc/calcpb"
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
		err := cc.Close()
		if err != nil {
			log.Fatalf("Cannot close a client connection: %v", err)
		}
	}(cc)

	cl := calcpb.NewCalcServiceClient(cc)
	makeSumCall(cl)
	makePrimeCall(cl)
	makeAverageCall(cl)
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
	numbers := []int{1, 2, 3, 4}

	stream, err := cl.Average(context.Background())
	if err != nil {
		log.Fatalf("Error while calling Average RPC: %v", err)
	}

	for _, n := range numbers {
		if err := stream.Send(&calcpb.AverageRequest{Number: int32(n)}); err != nil {
			log.Fatalf("Error while sending a request to the stream: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from Average: %v", err)
	}
	log.Printf("Response from Average: %v", res.GetResult())
}
