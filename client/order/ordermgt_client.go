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
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"productinfo/client/common"
	pb "productinfo/client/ecommerce"
)

var (
	address  = "server:50051"
	hostname = "localhost"
	crtFile  = "certs/client.crt"
	keyFile  = "certs/client.key"
	caFile   = "certs/ca.crt"
)

func main() {
	// 1. Загружаем сертификат и ключ клиента
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load client certificate and key: %v", err)
	}

	// 2. Загружаем CA
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("Failed to append CA certificate")
	}

	// 3. Настраиваем TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   hostname,
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS12,
	}

	// 4. Создаём учётные данные TLS
	creds := credentials.NewTLS(tlsConfig)

	// 5. Подключаемся к серверу с mTLS
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Order client: did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)

	ctx := common.WithAuthMetadata(context.Background())
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	order, err := client.GetOrder(ctx, &wrapperspb.StringValue{Value: "102"})
	if err != nil {
		log.Fatalf("Could not get order: %v", err)
	}
	fmt.Printf("Order: %v\n", order.String())
}
