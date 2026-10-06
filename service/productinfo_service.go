package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"productinfo/db/sqlc"
	pb "productinfo/service/ecommerce"
)

func numericFromFloat32(v float32) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%g", v)); err != nil {
		return n, err
	}
	return n, nil
}

func float32FromNumeric(n pgtype.Numeric) (float32, error) {
	f, err := n.Float64Value()
	if err != nil {
		return 0, err
	}
	if !f.Valid {
		return 0, fmt.Errorf("price is null")
	}
	return float32(f.Float64), nil
}

type server struct {
	pb.UnimplementedProductInfoServer
	queries *sqlc.Queries
}

func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	price, err := numericFromFloat32(in.Price)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid price: %v", err)
	}

	product, err := s.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		Name:        in.Name,
		Description: in.Description,
		Price:       price,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create product: %v", err)
	}

	return &pb.ProductID{Value: product.ID.String()}, nil
}

func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	productID, err := uuid.Parse(in.Value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid product ID: %v", err)
	}

	product, err := s.queries.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "Product not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get product: %v", err)
	}

	price, err := float32FromNumeric(product.Price)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to read product price: %v", err)
	}

	return &pb.Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       price,
	}, nil
}
