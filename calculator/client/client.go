package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/jwfrizzell/grpc-go-course/calculator/calculatorpb"

	"google.golang.org/grpc"
)

type options struct{}

func main() {
	fmt.Println("Initializing Client Connection...")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer cc.Close()
	if err != nil {
		log.Fatalf("Client Connection Failure: %v\n", err)
	}

	c := calculatorpb.NewCalculatorServiceClient(cc)

	// runUnary(c)

	// doServiceStreaming(c)

	//doClientStreaming(c)

	//doBiDirectionalStreaming(c)

	doUnaryErrorHandling(c, 10)

	doUnaryErrorHandling(c, -10)

}

func runUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting Unary RPC...")
	calc := &calculatorpb.Calculator{
		FirstNumber: 3,
		LastNumber:  10,
	}

	cr := &calculatorpb.CalculatorRequest{
		Calculator: calc,
	}
	resp, err := c.CalculateSum(context.Background(), cr)
	if err != nil {
		log.Fatalf("Calculate Sum Failure: %v\n", err)
	}
	log.Printf("Sum From Response: %v\n", resp.Result)

}

func doServiceStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting Server Side Streaming Server...")

	prime := &calculatorpb.Prime{
		Number: 120,
	}

	req := &calculatorpb.PrimeRequest{
		Prime: prime,
	}

	rs, err := c.CalculatePrimeDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Calculate Prime Number Error: %v", err)
	}

	for {
		msg, err := rs.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Prime Decomposition Stream Failure: %v", err)
		}
		log.Printf("Prime Response: %d", msg.Number)
	}

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting Client Side Streaming...")

	ca := []int64{3, 5, 9, 54, 23}
	stream, err := c.CalculateAverage(context.Background())
	if err != nil {
		log.Fatalf("doClientStreaming() Failure: %v", err)
	}

	for _, v := range ca {
		fmt.Println("Sending Client Streaming Request: ", v)
		err = stream.Send(&calculatorpb.AverageRequest{
			Number: v,
		})
		if err != nil {
			log.Fatalf("doClientStreaming() Failure: %v", err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("doClientStreaming() Failure: %v", err)
	}
	fmt.Printf("Average: %v\n", res.GetNumber())

}

func doBiDirectionalStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting Bi-Directional Client Side Streaming...")

	values := []float64{1, 5, 3, 6, 2, 20}

	stream, err := c.FindMax(context.Background())
	if err != nil {
		log.Fatalf("Bi-Directional Stream Error: %v", err)
		return
	}

	waitc := make(chan struct{})
	//Sending data to server.
	go func() {
		for _, v := range values {
			err := stream.Send(&calculatorpb.FindMaxRequest{
				Number: v,
			})
			if err != nil {
				log.Fatalf("doBiDirectionalStreaming Error: %v", err)
			}
			time.Sleep(time.Second)
		}
		err = stream.CloseSend()
		if err != nil {
			log.Fatalf("CloseSend() Error: %v", err)
		}
	}()

	//Receive Response
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("End of data stream...")
				break
			}
			if err != nil {
				log.Fatalf("Receive Response Error: %v", err)
				break
			}
			fmt.Printf("Server Response Value: %v\n", resp.GetNumber())
		}
		close(waitc)
	}()

	<-waitc
}

func doUnaryErrorHandling(c calculatorpb.CalculatorServiceClient, value float64) {
	fmt.Println("Starting Unary Error handling...")

	sr := &calculatorpb.SquareRootRequest{
		Number: value,
	}

	resp, err := c.SquareRoot(context.Background(), sr)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			//Actual error from GRPC
			fmt.Printf("GRPC Error: %v\nCode: %v\n", s.Message(), s.Code())
			if s.Code() == codes.InvalidArgument {
				fmt.Println("Bad Value Sent: ", sr.GetNumber())
			}
		} else {
			log.Fatalf("Undefined Error: %v", err)
		}
		return
	}
	fmt.Printf("Square Root Response: %v\n", resp.GetNumber())

}
