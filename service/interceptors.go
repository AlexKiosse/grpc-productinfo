package main

import (
	"context"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Перехватчик для логирования
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Логируем запрос (до вызова метода)
	log.Printf("[%s] method call: %s", time.Now().Format("15:04:05"), info.FullMethod)
	log.Printf("request: %+v", req)

	// Вызываем метод
	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		log.Printf("[%s] ERROR: %v (lead time: %v)", info.FullMethod, err, duration)
	} else {
		log.Printf("[%s] SUCCESSFULLY (lead time: %v)", info.FullMethod, duration)
		log.Printf("answer: %+v", resp)
	}
	return resp, err
}

// Перехватчик для авторизации
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Достаём метаданные из контекста (их прислал клиент)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		//Если метаданных нет - клиент не авторизован
		return nil, status.Error(codes.Unauthenticated, "Metadata not found. Client not authorized.")
	}

	// Проверяем наличие заголовка "authorization"
	authHeader := md["authorization"]
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "The Authorization header is missing.")
	}

	// Разбираем заголовок: должен быть в формате "Bearer <token>"
	tokenParts := strings.Split(authHeader[0], " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, status.Error(codes.Unauthenticated, "Invalid Authorization header format. Expected: Bearer <token>")
	}

	token := tokenParts[1]
	// Проверяем токен
	if !strings.HasPrefix(token, "eyJ") {
		return nil, status.Error(codes.PermissionDenied, "Invalid token. Access denied.")
	}

	// Логируем кто делает запрос
	requestID := md["x-request-id"]
	if len(requestID) > 0 {
		log.Printf("Request-ID: %s", requestID[0])
	}

	clientType := md["x-client-type"]
	if len(clientType) > 0 {
		log.Printf("Client-type: %s", clientType[0])
	}

	log.Printf(" Authorization successful. Token: %s...", token[:20])

	// Всё хорошо - вызываем метод
	return handler(ctx, req)
}

// Перехватчик для обработки PANIC
func recoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in method %s: %v", info.FullMethod, r)
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
