package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "proto-to-mcp-tutorial/generated/go"
)

type server struct {
	pb.UnimplementedBookstoreServiceServer
	books map[string]*pb.Book
}

func (s *server) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.Book, error) {
	book, exists := s.books[req.BookId]
	if !exists {
		return nil, fmt.Errorf("book not found")
	}
	return book, nil
}

func (s *server) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.Book, error) {
	switch {
	case req.Book == nil:
		return nil, fmt.Errorf("book is required")
	case req.Book.Title == "":
		return nil, fmt.Errorf("book title is required")
	case req.Book.Author == "":
		return nil, fmt.Errorf("book author is required")
	case req.Book.Pages <= 0:
		return nil, fmt.Errorf("book pages must be greater than zero")
	}

	book := req.Book

	fmt.Println("Creating book:", book.Author, book.Title, book.Pages)
	if book.BookId == "" {
		book.BookId = fmt.Sprintf("book-%d", len(s.books)+1)
	}
	s.books[book.BookId] = book
	return book, nil
}

func main() {
	// Initialize server with some sample data
	srv := &server{
		books: map[string]*pb.Book{
			"book-1": {
				BookId: "book-1",
				Title:  "The Go Programming Language",
				Author: "Alan Donovan",
				Pages:  380,
			},
		},
	}

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterBookstoreServiceServer(s, srv)

		reflection.Register(s)

		log.Println("gRPC server starting on :9090")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start REST gateway
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterBookstoreServiceHandlerFromEndpoint(ctx, mux, "localhost:9090", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	log.Println("REST server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve REST: %v", err)
	}
}
