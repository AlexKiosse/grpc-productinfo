package main

import (
	"context"
	"fmt"
	"log"
	pb "productinfo/client/ecommerce"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.UpdateOrders(ctx)
	if err != nil {
		log.Fatalf("Error calling UpdateOrders: %v", err)
	}

	ordersToUpdate := []*pb.Order{
		{Id: "100", Description: "Updated order 100"},
		{Id: "101", Description: "Updated order 101"},
		{Id: "102", Description: "Updated order 102"},
	}

	for _, order := range ordersToUpdate {
		if err := stream.Send(order); err != nil {
			log.Fatalf("Error sending order: %v", err)
		}
		log.Printf("Sent order %s for update: ", order.Id)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error receiving response: %v", err)
	}
	fmt.Printf("Server response: %s\n", response.Value)
}
