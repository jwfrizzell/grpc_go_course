package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/jwfrizzell/grpc-go-course/blog/blogpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Staring Blog Client...")

	//Creating Client
	host := "localhost:50051"
	cc, err := grpc.Dial(host, grpc.WithInsecure())
	defer cc.Close()
	if err != nil {
		log.Fatalf("Unable to Dial Host: %s.\n", host)
		return
	}
	client := blogpb.NewBlogServiceClient(cc)
	log.Printf("Client Created...\n")

	//These are just examples and use hardcoded values.
	//The IDs need to be updated when using.
	createBlog(client)

	readBlog(client)

	updateBlog(client)

	//deleteBlog(client)

	listAllBlogs(client)

}

func readBlog(client blogpb.BlogServiceClient) {
	log.Println("Client Calling readBlog()...")

	resp, err := client.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: "5cdb2104ebccb611c68d989a", //Replace this with existing ID in blog database
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	blog := resp.GetBlog()
	log.Println("Blog ID: ", blog.GetId())
	log.Println("Author: ", blog.GetAuthorId())
	log.Println("Title: ", blog.GetTitle())
	log.Println("Content: ", blog.GetContent())
}

func createBlog(client blogpb.BlogServiceClient) {
	log.Println("Client Calling createBlog()...")
	//Creating Blog
	resp, err := client.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "Joe Frizzell",
			Content:  "Personal Blog!",
			Title:    "Title for my very own blog.",
		},
	})
	if err != nil {
		log.Fatalf("Server Response Error: %v\n", err)
		return
	}
	fmt.Println(resp.GetBlog())
}

func updateBlog(client blogpb.BlogServiceClient) {
	log.Println("Client Calling updateBlog()...")

	req := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       "5cdb2104ebccb611c68d989a", //Replace this with existing ID in blog database
			AuthorId: "Joe Frizzell III",
			Title:    "How To Update a Blog...",
			Content:  "This is me updating my blog like a boss!",
		},
	}

	resp, err := client.UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatal(err)
		return
	}

	blog := resp.GetBlog()
	log.Println("Blog ID: ", blog.GetId())
	log.Println("Author: ", blog.GetAuthorId())
	log.Println("Title: ", blog.GetTitle())
	log.Println("Content: ", blog.GetContent())
}

func deleteBlog(client blogpb.BlogServiceClient) {
	log.Println("Client Calling updateBlog()...")

	ids := []string{
		"5cdb2140ebccb611c68d989b",
	} //Replace this with existing ID in blog database

	for _, v := range ids {
		log.Println("Removing ID: ", v)
		req := &blogpb.DeleteBlogRequest{
			BlogId: v,
		}

		res, err := client.DeleteBlog(context.Background(), req)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("Deleted BlogID: %s\n", res.GetBlogId())
	}
}

func listAllBlogs(client blogpb.BlogServiceClient) {
	log.Println("Client Calling listAllBlogs()...")

	stream, err := client.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer stream.CloseSend()
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			return
		}
		b := res.GetBlog()
		fmt.Printf("\n\nBlogID: %s\nAuthor: %s\nTitle: %s\nContent: %s",
			b.GetId(), b.GetAuthorId(), b.GetTitle(), b.GetContent())
	}

}
