package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jwfrizzell/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
)

type server struct{}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Invoking Greet() function. Request: %v\n", req)

	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	result := fmt.Sprintf("Hello %s %s", firstName, lastName)

	response := &greetpb.GreetResponse{
		Result: result,
	}
	return response, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("Invoking GreetManyTimes() function. Request: %v\n", req)

	firstName := req.GetGreeting().GetLastName()
	lastName := req.GetGreeting().GetLastName()

	for i := 1; i <= 10; i++ {

		resp := &greetpb.GreetManyTimesResponse{
			Result: fmt.Sprintf("Hello %s %s. You are number %d.", firstName, lastName, i),
		}
		stream.Send(resp)
		time.Sleep(time.Second)
	}
	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Printf("Invoking LongGreet() function...")
	result := "Hi %s %s! "
	var a string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: a,
			})
		}
		if err != nil {
			log.Fatalf("LongGreet Stream Failure: %v", err)
		}

		fn := req.GetGreeting().GetFirstName()
		ln := req.GetGreeting().GetLastName()

		a += fmt.Sprintf(result, fn, ln)
	}
}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("Invoking GreetEveryone() function...")
	fs := "Hello %s %s! "
	for {
		ger, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Receive Stream Failure: %v", err)
			return err
		}
		result := fmt.Sprintf(fs, ger.GetGreeting().GetFirstName(), ger.GetGreeting().GetLastName())
		err = stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if err != nil {
			log.Fatalf("Error when sending streaming data to client. Error: %v", err)
			return err
		}

	}
}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Println("Invoking GreetWithDealine() function...")

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			//Client cancelled request.
			fmt.Println("Client cancelled request...")
			return nil, status.Error(codes.DeadlineExceeded, "Client has cancelled the request.")
		}
		time.Sleep(time.Second)
	}

	fn := req.GetGreeting().GetFirstName()
	ln := req.GetGreeting().GetLastName()

	resp := &greetpb.GreetWithDeadlineResponse{
		Result: fmt.Sprintf("Hello %s %s", fn, ln),
	}
	return resp, nil
}

func main() {
	fmt.Println("GRPC Server has started...")

	tls := true
	opts := []grpc.ServerOption{}
	if tls {
		certfile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certfile, keyFile)
		if err != nil {
			log.Fatalf("Unable to create New Client TLS From File. SSL Error: %v", err)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}
	s := grpc.NewServer(opts...)
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Listen Failure: %s", err)
	}

	greetpb.RegisterGreetServiceServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Serve Failure: %s", err)
	}
	fmt.Println("Server running...")

}
