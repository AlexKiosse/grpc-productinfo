package main

import (
	"context"
	"os"
	"productinfo/db/sqlc"
	pb "productinfo/service/ecommerce"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = ":50051"
)

func certPath(name string) string {
	dir := os.Getenv("CERT_DIR")
	if dir == "" {
		dir = "../certs"
	}
	return filepath.Join(dir, name)
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL environment variable is not set")
	}

	// Пул соединения с БД
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database is not reachable: %v", err)
	}
	log.Println("Connect to database")

	// Для работы с БД
	queries := sqlc.New(pool)

	srv := &server{
		queries: queries,
	}

	// Загружаем ключ и сертификат сервера
	certificate, err := tls.LoadX509KeyPair(certPath("server.crt"), certPath("server.key"))
	if err != nil {
		log.Fatalf("Failed to load the server certificate and key: %v", err)
	}

	// Создаём пул доверенных сертификатов для проверки клиентов
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(certPath("ca.crt"))
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

	pb.RegisterProductInfoServer(s, srv)
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
