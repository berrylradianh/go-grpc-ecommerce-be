package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/handler"
	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/repository"
	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/service"
	"github.com/berrylradianh/go-grpc-ecommerce-be/pb/auth"
	"github.com/berrylradianh/go-grpc-ecommerce-be/pkg/database"
	"github.com/berrylradianh/go-grpc-ecommerce-be/pkg/grpcmiddleware"
	"github.com/joho/godotenv"
	gocache "github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Panicf("failed to listen: %v", err)
	}

	db := database.ConnectDB(ctx, os.Getenv("DB_URI"))
	log.Println("Database connected")

	cacheService := gocache.New(time.Hour*24, time.Hour)

	authRepository := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepository, cacheService)
	authHandler := handler.NewAuthHandler(authService)

	serv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.ErrorMiddleware,
		),
	)

	auth.RegisterAuthServiceServer(serv, authHandler)

	if os.Getenv("ENVIRONMENT") == "dev" {
		reflection.Register(serv)
		log.Println("Reflection service registered")
	}

	log.Printf("Starting server on port %v", lis.Addr())
	if err := serv.Serve(lis); err != nil {
		log.Panicf("failed to serve: %v", err)
	}
}
