package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jwfrizzell/grpc-go-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
)

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

var collection *mongo.Collection

//Server Entry Point
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting Mongodb...")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Mongodb Client Error: %v\n", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Mongodb Connection Error: %v\n", err)
		return
	}
	collection = client.Database("blogdb").Collection("blog")
	log.Println("Mongodb has been successfully started...")

	log.Println("Staring Blog Servcie.")
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Server Listen Error: %v\n", err)
		return
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)

	blogpb.RegisterBlogServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		log.Println("Staring Blog Server.")
		if err := s.Serve(lis); err != nil {
			log.Fatal("Unable to serve connections on listener.\n")
			return
		}

	}()

	//Wait for Control C
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until signal is received.
	<-ch
	log.Println("\nDisconnecting Mongodb Client")
	client.Disconnect(ctx)
	log.Println("Stopping Blog Server")
	s.Stop()
	log.Println("Closing Listener")
	lis.Close()

}

//Create Unary Blog Request
func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Println("Starting CreateBlog Server Request...")

	blog := req.GetBlog()

	//ID,AuthorID,Content,Title
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal Error: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Cannot Convert OID")
	}

	br := &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}
	return br, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Println("Starting ReadBlog Server Request...")

	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot Parse ID!")
	}

	//Create empty strucxt
	data := &blogItem{}
	filter := bson.M{"_id": oid}
	sr := collection.FindOne(context.Background(), filter)
	if err := sr.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find Blog ID %v", oid))
	}

	resp := &blogpb.ReadBlogResponse{
		Blog: dataToBlogPB(data),
	}
	return resp, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	log.Println("Starting UpdateBlog Server Request...")

	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Unable to Parse ID.")
	}

	filter := bson.M{
		"_id": oid,
	}
	data := &blogItem{}
	sr := collection.FindOne(context.Background(), filter)
	if err := sr.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find Blog ID %v", oid))
	}

	//We update our internal struct.
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, err = collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to update Blog: %v", oid)
	}

	resp := &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPB(data),
	}
	return resp, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	log.Println("Starting DeleteBlog Server Request...")

	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Error Parsing ID.")
	}
	filter := bson.M{
		"_id": oid,
	}
	dr, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error Deleting Record: %s", oid))
	}
	if dr.DeletedCount == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Record Not Found: %s", oid))
	}
	log.Printf("Number or records deleted %d", dr.DeletedCount)

	res := &blogpb.DeleteBlogResponse{
		BlogId: oid.Hex(),
	}
	return res, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	log.Println("Starting ListBlog Server Request...")

	cur, err := collection.Find(context.TODO(), nil)
	if err != nil {
		return status.Error(codes.NotFound, fmt.Sprintf("Unable to get list from MongoDB. Error: %v", err))
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &blogItem{}
		if err = cur.Decode(data); err != nil {
			return status.Error(codes.Internal, "Unable to retrieve data from cursor.")
		}
		err = stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlogPB(data),
		})
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Could create response from cursor. Error: %v", err))
		}
	}
	if err = cur.Err(); err != nil {
		return status.Error(codes.Unknown, fmt.Sprintf("Error: %v", err))
	}
	return nil
}

func dataToBlogPB(data *blogItem) *blogpb.Blog {

	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}
