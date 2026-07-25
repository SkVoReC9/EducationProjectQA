package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	// Импорты сгенерированных контрактов
	pbCart "awesomeProject/gen/store/api/cart/v1"
	pbCatalog "awesomeProject/gen/store/api/catalog/v1"
	pbOrder "awesomeProject/gen/store/api/order/v1"

	"awesomeProject/internal/handler"
	"awesomeProject/internal/repository/postgres"
	"awesomeProject/internal/service"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://store:store@localhost:5432/store?sslmode=disable"
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// --- 1. Инициализация Каталога ---
	catalogRepo := postgres.NewCatalogRepository(db)
	catalogService := service.NewCatalogService(catalogRepo)
	catalogHandler := handler.NewCatalogHandler(catalogService)

	// --- 2. Инициализация Корзины ---
	cartRepo := postgres.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo, catalogService)
	cartHandler := handler.NewCartHandler(cartService)

	// --- 3. Инициализация Заказов ---
	orderRepo := postgres.NewOrderRepository(db)
	// Обрати внимание: OrderService забирает себе CartService и CatalogService
	orderService := service.NewOrderService(orderRepo, cartService, catalogService)
	orderHandler := handler.NewOrderHandler(orderService)

	// --- Настройка gRPC сервера ---
	grpcPort := ":50051"
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// --- Регистрация ВСЕХ хэндлеров ---
	pbCatalog.RegisterCatalogServiceServer(grpcServer, catalogHandler)
	pbCart.RegisterCartServiceServer(grpcServer, cartHandler)
	pbOrder.RegisterOrderServiceServer(grpcServer, orderHandler)

	// Рефлексия подхватит все зарегистрированные выше сервисы
	reflection.Register(grpcServer)

	go func() {
		fmt.Printf("Starting gRPC Simulator on %s...\n", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// --- HTTP gateway (JSON REST → gRPC) ---
	httpPort := ":8080"
	if err := runHTTPGateway(context.Background(), grpcPort, httpPort); err != nil {
		log.Fatalf("failed to serve HTTP gateway: %v", err)
	}
}

func runHTTPGateway(ctx context.Context, grpcEndpoint, httpPort string) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pbCatalog.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register catalog gateway: %w", err)
	}
	if err := pbCart.RegisterCartServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register cart gateway: %w", err)
	}
	if err := pbOrder.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("register order gateway: %w", err)
	}

	fmt.Printf("Starting HTTP gateway on %s...\n", httpPort)
	return http.ListenAndServe(httpPort, mux)
}
