package common

import (
	"context"
	"time"

	"google.golang.org/grpc/metadata"
)

// Функция создаёт контекст с метаданными
func WithAuthMetadata(ctx context.Context) context.Context {
	fakeToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	requestID := "req-" + time.Now().Format("20060102150405")

	// Создаёт метаданные
	md := metadata.Pairs(
		"authorization", fakeToken, // Токен доступа
		"x-request-id", requestID, // ID запроса
		"x-client-type", "go-cli", // Тип клиента
	)

	// Добавляем метаданные в контекст
	return metadata.NewOutgoingContext(ctx, md)
}
