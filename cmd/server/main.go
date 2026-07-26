package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pbCart "awesomeProject/gen/store/api/cart/v1"
	pbCatalog "awesomeProject/gen/store/api/catalog/v1"
	pbOrder "awesomeProject/gen/store/api/order/v1"
	pbPromo "awesomeProject/gen/store/api/promo/v1"
	pbUser "awesomeProject/gen/store/api/user/v1"

	"awesomeProject/internal/auth"
	"awesomeProject/internal/handler"
	"awesomeProject/internal/repository/postgres"
	"awesomeProject/internal/service"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://store:store@localhost:5432/store?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}
	jwtManager := auth.NewManager(jwtSecret, 24*time.Hour)

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	userService := service.NewUserService(userRepo, jwtManager)
	userHandler := handler.NewUserHandler(userService)

	catalogRepo := postgres.NewCatalogRepository(db)
	catalogService := service.NewCatalogService(catalogRepo)
	catalogHandler := handler.NewCatalogHandler(catalogService)

	promoRepo := postgres.NewPromoRepository(db)
	promoService := service.NewPromoService(promoRepo)
	promoHandler := handler.NewPromoHandler(promoService)

	cartRepo := postgres.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo, catalogService, userService, promoService)
	cartHandler := handler.NewCartHandler(cartService)

	orderRepo := postgres.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, cartService, catalogService, userService)
	orderHandler := handler.NewOrderHandler(orderService)

	grpcPort := ":50051"
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryInterceptor(jwtManager)),
	)

	pbUser.RegisterUserServiceServer(grpcServer, userHandler)
	pbCatalog.RegisterCatalogServiceServer(grpcServer, catalogHandler)
	pbCart.RegisterCartServiceServer(grpcServer, cartHandler)
	pbOrder.RegisterOrderServiceServer(grpcServer, orderHandler)
	pbPromo.RegisterPromoServiceServer(grpcServer, promoHandler)

	reflection.Register(grpcServer)

	go func() {
		fmt.Printf("Starting gRPC Simulator on %s...\n", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	httpPort := ":8080"
	if err := runHTTPGateway(context.Background(), grpcPort, httpPort); err != nil {
		log.Fatalf("failed to serve HTTP gateway: %v", err)
	}
}

func runHTTPGateway(ctx context.Context, grpcEndpoint, httpPort string) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pbUser.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register user gateway: %w", err)
	}
	if err := pbCatalog.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register catalog gateway: %w", err)
	}
	if err := pbCart.RegisterCartServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register cart gateway: %w", err)
	}
	if err := pbOrder.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register order gateway: %w", err)
	}
	if err := pbPromo.RegisterPromoServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register promo gateway: %w", err)
	}

	fmt.Printf("Starting HTTP gateway on %s...\n", httpPort)
	return http.ListenAndServe(httpPort, mux)
}
