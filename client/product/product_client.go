package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"productinfo/client/common"
	pb "productinfo/client/ecommerce"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
