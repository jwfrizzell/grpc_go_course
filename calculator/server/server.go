package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc/codes"

	"github.com/jwfrizzell/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type server struct{}

func (s *server) SquareRoot(CTX context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Invoking SquareRoot() Function...")

	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument,
			fmt.Sprintf("Received a negative number %v.", number))
	}
	resp := &calculatorpb.SquareRootResponse{
		Number: math.Sqrt(number),
	}
	return resp, nil
}

func (s *server) CalculateSum(ctx context.Context, req *calculatorpb.CalculatorRequest) (*calculatorpb.CalculatorResponse, error) {
	fmt.Println("Invoking CalculateSum() Function...")

	fn := req.GetCalculator().GetFirstNumber()
	ln := req.GetCalculator().GetLastNumber()
	sum := fn + ln

	resp := &calculatorpb.CalculatorResponse{
		Result: sum,
	}

	return resp, nil
}

func (s *server) CalculatePrimeDecomposition(req *calculatorpb.PrimeRequest, stream calculatorpb.CalculatorService_CalculatePrimeDecompositionServer) error {
	fmt.Println("Invoking CalculatePrimeDecomposition()...")

	n := req.GetPrime().GetNumber()
	k := int64(2)

	for n > 1 {
		if n%k == 0 {
			stream.Send(&calculatorpb.PrimeResponse{
				Number: k,
			})
			n = n / k
		} else {
			k++
			fmt.Println("Divisor: ", k)
		}
	}
	return nil
}

func (s *server) CalculateAverage(stream calculatorpb.CalculatorService_CalculateAverageServer) error {
	fmt.Println("Invoking CalculateAverage()...")

	var a int64
	var i float64
	for {
		rec, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Number: float64(a) / i,
			})
		}
		if err != nil {
			log.Fatalf("CalculateAverage() Client Stream Failure: %v\n", err)
		}
		a += rec.GetNumber()
		i++
		fmt.Printf("Number: %d - Iteration: %f\n", a, i)
	}
}

func (s *server) FindMax(stream calculatorpb.CalculatorService_FindMaxServer) error {
	fmt.Println("Invoking FindMax() Server...")

	max := float64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("FindMax Recv Error: %v", err)
			return err
		}

		n := req.GetNumber()
		fmt.Println("Numbers sent from client: ", n)
		if n > max {
			max = n
			err = stream.Send(&calculatorpb.FindMaxResponse{
				Number: max,
			})
			if err != nil {
				log.Fatalf("FindMax Send Error: %v", err)
				return err
			}
		}

	}
}

func main() {
	fmt.Println("Starting Client Server...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Server Failure: %v\n", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Server Listen Failure: %v\n", err)
	}
	fmt.Println("Server is running...")

}
