package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"prodApp/proto/gen/notifierpb"
	srv "prodApp/tg-notifier/internal/transport/grpc"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
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

	log.Printf("tg-notifier gRPC server started on :%s", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpc server failed: %v", err)
	}
}
