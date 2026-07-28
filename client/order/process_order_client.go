package main

import (
	"context"
	"log"
	pb "productinfo/client/ecommerce"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("Error calling ProcessOrder: %v", err)
	}

	orderIDs := []string{"100", "101", "102", "100", "100", "101", "102"}

	done := make(chan bool)

	go func() {
		for {
			shipment, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving shipment: %v", err)
				break
			}
			log.Printf("Received shipment: ID=%s, Status=%v, Orders=%d", shipment.Id, shipment.Status, len(shipment.OrderList))
			for _, order := range shipment.OrderList {
				log.Printf(" Order: %s -> %s", order.Id, order.Destination)
			}
		}
		done <- true
	}()

	for _, id := range orderIDs {
		if err := stream.Send(&wrapperspb.StringValue{Value: id}); err != nil {
			log.Fatalf("Error sending order ID: %v", err)
		}
		log.Printf("Sent order ID: %s", id)
		time.Sleep(500 * time.Millisecond)
	}

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("Error closing send stream %v", err)
	}

	<-done
	log.Println("ProcessOrders completed")
}
