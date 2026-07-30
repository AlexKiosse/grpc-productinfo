package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"productinfo/client/common"
	pb "productinfo/client/ecommerce"
)

var (
	hostname = "localhost"
	crtFile  = "certs/client.crt"
	keyFile  = "certs/client.key"
	caFile   = "certs/ca.crt"
)

func main() {
	// Загружаем ключ и сертификат клиента
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load the server certificate and key: %v", err)
	}

	// Создаём пул доверенных сертификатов для проверки сервера
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Failed to read the CA certificate: %v", err)
	}

	// Добавляем CA в пул
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("Failed to add the CA to the pool")
	}

	// Настраиваем TLS для клиента
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   hostname,
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS12,
	}

	// Создаём учётные данные TLS
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.Dial("server:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Product client: did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewProductInfoClient(conn)

	// Создаём контекст с метаданными
	ctx := common.WithAuthMetadata(context.Background())

	// Добавляем крайний срок
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Добавляем товар
	log.Println("Отправляем запрос на добавление товара...")

	r, err := client.AddProduct(ctx, &pb.Product{
		Name:        "Apple iPhone 17",
		Description: "Meet Apple iPhone 17. All-new dual-camera system with Ultra Wide and Night mode.",
		Price:       2000.0,
	})
	// Обработка ошибки и возращаем норм ошибку
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			log.Fatalf("The deadline has expired.")
		}
		log.Fatalf("Error adding product: %v", err)
	}
	fmt.Printf("Product ID: %s added successfully\n", r.Value)

	// Получаем товар
	product, err := client.GetProduct(ctx, &pb.ProductID{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}
	fmt.Printf("Product: %v\n", product.String())
}
