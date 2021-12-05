package main

import (
	"context"
	"github.com/lozhkindm/grpc-go/calc/calcpb"
	"google.golang.org/grpc"
	"io"
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

func (server) Average(stream calcpb.CalcService_AverageServer) error {
	sum, num := .0, 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := sum / float64(num)
			return stream.SendAndClose(&calcpb.AverageResponse{Result: res})
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		sum += float64(req.GetNumber())
		num++
	}
}

func (server) Maximum(stream calcpb.CalcService_MaximumServer) error {
	max := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if num := req.GetNumber(); num > max {
			max = num
			if err := stream.Send(&calcpb.MaximumResponse{Result: max}); err != nil {
				return err
			}
		}
	}
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
