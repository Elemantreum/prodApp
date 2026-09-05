package main

import (
	"log"
	"net"
	"os"
	"prodApp/proto/gen/notifierpb"
	srv "prodApp/vk-notifier/internal/transport/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	token := os.Getenv("VK_APP_SECRET_KEY")
	if token == "" {
		log.Fatal("VK_APP_SECRET_KEY is not set")
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	notifierServer := srv.NewServer(token)

	notifierpb.RegisterNotifierServiceServer(
		grpcServer,
		notifierServer,
	)

	reflection.Register(grpcServer)

	log.Printf("vk-notifier gRPC server started on :%s", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpc server failed: %v", err)
	}
}
