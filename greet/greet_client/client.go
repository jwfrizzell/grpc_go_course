package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/jwfrizzell/grpc-go-course/greet/greetpb"

	"google.golang.org/grpc"
)

type options struct{}

func main() {
	fmt.Println("Establishing Client Connection...")

	tls := true
	opts := grpc.WithInsecure()
	if tls {
		creds, err := credentials.NewClientTLSFromFile("ssl/ca.crt", "")
		if err != nil {
			log.Fatalf("Unable to create client from file. Error: %v", err)
			return
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)
	defer cc.Close()
	if err != nil {
		log.Fatalf("Connection Failed: %s", err)
	}

	c := greetpb.NewGreetServiceClient(cc)

	doUnary(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	//doBiDirectionalStreaming(c)

	// doUnaryWithDeadline(c, time.Second*5)
	// doUnaryWithDeadline(c, time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting Unary RPC...")
	g := &greetpb.Greeting{
		FirstName: "Joe",
		LastName:  "Frizzell",
	}

	gr := &greetpb.GreetRequest{
		Greeting: g,
	}

	resp, err := c.Greet(context.Background(), gr)
	if err != nil {
		log.Fatalf("GreetingResponse Failed: %v\n", err)
	}
	log.Printf("Greeting from Response: %v\n", resp.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting Server Streaming RPC...")

	gr := &greetpb.Greeting{
		FirstName: "Joe",
		LastName:  "Frizzell",
	}
	gmt := &greetpb.GreetManyTimesRequest{
		Greeting: gr,
	}

	rs, err := c.GreetManyTimes(context.Background(), gmt)
	if err != nil {
		log.Fatalf("GreetManyTimesResponse Failed: %v\n", err)
	}
	for {
		msg, err := rs.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("GreetManyTimesResponse Receive Failed: %v\n", err)
		}
		log.Printf("Greeting from GMT Response: %v\n", msg.Result)
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Start Client Streaming RPC...")

	lg := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Joe",
				LastName:  "Frizzell",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Cullin",
				LastName:  "Frizzell",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Krystal",
				LastName:  "Frizzell",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Erin",
				LastName:  "Frizzell",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "No",
				LastName:  "One",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shogo",
				LastName:  "Makishima",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("doClientStreaming() Failure: %v", err)
	}

	for _, v := range lg {
		fmt.Println("Sending Streaming Requests: ", v)
		err = stream.Send(v)
		if err != nil {
			log.Fatalf("doClientStreaming() Failure: %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("doClientStreaming() CloseAndRecv Failure: %v", err)
	}
	fmt.Printf("Client Streaming Response: %v\n", resp)
}

func doBiDirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Start Bidirectional Streaming...")

	lg := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Izia",
				LastName:  "Orihara",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tony",
				LastName:  "Stark",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Johann",
				LastName:  "Liebert",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shogo",
				LastName:  "Makishima",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Allison",
				LastName:  "Argent",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shogo",
				LastName:  "Makishima",
			},
		},
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error creating stream for client. Error: %v", err)
		return
	}

	waitc := make(chan struct{})
	//Send Requests to server.
	go func() {
		for _, v := range lg {
			fmt.Println("Sending message to server...")
			err := stream.Send(v)
			if err != nil {
				log.Fatalf("Send BI Streaming to Server: Error: %v", err)
			}
		}
		err = stream.CloseSend()
		if err != nil {
			log.Fatalf("Error closing stream. Error: %v", err)
		}
	}()

	//Receive Messages from server.
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error receiving response from server. Error: %v", err)
			}
			fmt.Println("Response: ", resp.Result)
		}
		close(waitc)
	}()
	<-waitc
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, time time.Duration) {
	fmt.Println("Start Dealine Unary Streaming...")

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Joe",
			LastName:  "Frizzell",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time)
	defer cancel()

	resp, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			fmt.Printf("GRPC Error\nMessage: %v\nCode:%s\n", s.Message(), s.Code())
			if s.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit and dealine was exceeded.")
			}
		} else {
			fmt.Println("Undefined Error: ", err)
		}
		return
	}
	fmt.Println(resp.GetResult())

}
