package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/calc/calcpb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct{}

func (s server) Sum(_ context.Context, req *calcpb.CalcRequest) (*calcpb.CalcResponse, error) {
	return &calcpb.CalcResponse{Result: req.Number1 + req.Number2}, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	calcpb.RegisterCalcServiceServer(srv, &server{})

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
