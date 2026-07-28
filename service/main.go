package main

import (
	pb "productinfo/service/ecommerce"

	"log"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Подключаем цепочку перехватчиков
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			recoveryInterceptor, // зашита от падений
			authInterceptor,     // проверка авторизации
			loggingInterceptor,  // логирование
		)),
	)

	pb.RegisterProductInfoServer(s, &server{})
	pb.RegisterOrderManagementServer(s, &orderServer{})

	log.Printf("Starting gRPC listener on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
