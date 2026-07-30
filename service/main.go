package main

import (
	pb "productinfo/service/ecommerce"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = ":50051"
)

var (
	crtFile = "certs/server.crt"
	keyFile = "certs/server.key"
	caFile  = "certs/ca.crt"
)

func main() {
	// Загружаем ключ и сертификат сервера
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load the server certificate and key: %v", err)
	}

	// Создаём пул доверенных сертификатов для проверки клиентов
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Failed to read the CA certificate: %v", err)
	}

	// Добавляем CA в пул
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("Failed to add the CA to the pool")
	}

	// Настройка mTLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	}

	// Создаём учётные данные TLS
	creds := credentials.NewTLS(tlsConfig)

	// Создаём grpc-сервер с mTLS и подключаем цепочку перехватчиков
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			recoveryInterceptor, // зашита от падений
			authInterceptor,     // проверка авторизации
			loggingInterceptor,  // логирование
		)),
		grpc.Creds(creds),
	)

	pb.RegisterProductInfoServer(s, &server{})
	pb.RegisterOrderManagementServer(s, &orderServer{})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Starting gRPC listener on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
