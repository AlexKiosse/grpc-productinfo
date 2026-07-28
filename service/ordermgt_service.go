package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	pb "productinfo/service/ecommerce"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type orderServer struct {
	pb.UnimplementedOrderManagementServer
}

var orderMap = map[string]*pb.Order{
	"100": {
		Id:          "100",
		Items:       []string{"Google Pixel 8 Pro", "Galaxy Buds Pro"},
		Description: "Order for electronics",
		Price:       2100.00,
		Destination: "Desnogorsk",
	},
	"101": {
		Id:          "101",
		Items:       []string{"Samsung S24 Ultra", "Galaxy Buds Pro 3"},
		Description: "Order for Samsung devices",
		Price:       2300,
		Destination: "Smolensk",
	},
	"102": {
		Id:          "102",
		Items:       []string{"Iphone 17 Pro", "Apple Watch 3", "AirPods 2"},
		Description: "Order for Apple devices",
		Price:       3000,
		Destination: "Moscow",
	},
}

// Унарный метод
func (s *orderServer) GetOrder(ctx context.Context, orderId *wrapperspb.StringValue) (*pb.Order, error) {
	// Проверка не отменил ли клиент запрос
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "The request was cancelled by the client or the deadline expired.")
	}
	ord, exists := orderMap[orderId.Value]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "Order does not exist: %s", orderId.Value)
	}
	return ord, nil

}

// Потоковый метод на стороне сервера
func (s *orderServer) SearchOrders(searchQuery *wrapperspb.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {
	for _, order := range orderMap {
		for _, itemStr := range order.Items {
			if strings.Contains(itemStr, searchQuery.Value) {
				err := stream.Send(order)
				if err != nil {
					return fmt.Errorf("error sending message to stream: %v", err)
				}
				break
			}
		}
	}
	return nil
}

// Потоквый метод на стороне клиента
func (s *orderServer) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	var updateIDs []string

	for {
		order, err := stream.Recv()
		if err == nil {
			updateIDs = append(updateIDs, order.Id)
			log.Printf("Order %s received for update", order.Id)
			continue
		}

		if err == io.EOF {
			return stream.SendAndClose(&wrapperspb.StringValue{
				Value: fmt.Sprintf("Updated orders: %v", updateIDs),
			})
		}
		return err
	}
}

// Двунаправленный поток
func (s *orderServer) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	shipments := make(map[string][]*pb.Order)
	batchSize := 3

	for {
		orderID, err := stream.Recv()
		if err == io.EOF {
			return s.sendAllShipments(stream, shipments)
		}
		if err != nil {
			return err
		}

		order, exist := orderMap[orderID.Value]
		if !exist {
			log.Printf("Order %s not found, skipping", orderID.Value)
			continue
		}

		dest := order.Destination
		shipments[dest] = append(shipments[dest], order)

		if len(shipments[dest]) >= batchSize {
			if err := s.sendShipment(stream, dest, shipments[dest]); err != nil {
				return err
			}
			delete(shipments, dest)
		}
	}
}

func (s *orderServer) sendShipment(stream pb.OrderManagement_ProcessOrdersServer, dest string, orders []*pb.Order) error {
	shipment := &pb.CombinedShipment{
		Id:        fmt.Sprintf("shipment-%s-%d", dest, time.Now().UnixNano()),
		Status:    "PROCESSED",
		OrderList: orders,
	}
	log.Printf("Sending shipment: %s with %d orders", shipment.Id, len(orders))
	return stream.Send(shipment)
}

func (s *orderServer) sendAllShipments(stream pb.OrderManagement_ProcessOrdersServer, shipments map[string][]*pb.Order) error {
	for dest, orders := range shipments {
		if err := s.sendShipment(stream, dest, orders); err != nil {
			return err
		}
	}
	return nil
}
