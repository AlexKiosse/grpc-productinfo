package main

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "productinfo/client/ecommerce"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect %v", err)
	}
	defer conn.Close()

	c := pb.NewOrderManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	searchStream, err := c.SearchOrders(ctx, &wrapperspb.StringValue{Value: "Google"})
	if err != nil {
		log.Fatalf("Error calling SearchOrders: %v", err)
	}

	log.Println("Search for orders containing 'Google'")

	for {
		searchOrder, err := searchStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error receiving: %v", err)
		}
		log.Printf("Search Result: %v", searchOrder)
	}
	log.Println("Search is completed")
}
