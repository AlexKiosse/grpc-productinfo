package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"productinfo/client/common"
	pb "productinfo/client/ecommerce"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Order client: did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)

	ctx := common.WithAuthMetadata(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	order, err := client.GetOrder(ctx, &wrapperspb.StringValue{Value: "102"})
	if err != nil {
		log.Fatalf("Could not get order: %v", err)
	}
	fmt.Printf("Order: %v\n", order.String())
}
