package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/calc/calcpb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct{}

func (server) Sum(_ context.Context, req *calcpb.CalcRequest) (*calcpb.CalcResponse, error) {
	return &calcpb.CalcResponse{Result: req.Number1 + req.Number2}, nil
}

func (server) Prime(req *calcpb.PrimeRequest, stream calcpb.CalcService_PrimeServer) error {
	div, num := 2, int(req.GetNumber())

	for num > 1 {
		if num%div == 0 {
			res := &calcpb.PrimeResponse{Result: int32(div)}
			if err := stream.Send(res); err != nil {
				return err
			}
			num /= div
		} else {
			div++
		}
	}

	return nil
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
