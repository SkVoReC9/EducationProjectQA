package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Импортируем сгенерированный код контрактов
	pb "awesomeProject/gen/store/api/catalog/v1"

	"awesomeProject/internal/repository"
	"awesomeProject/internal/service"
)

// CatalogHandler реализует сгенерированный интерфейс CatalogServiceServer
type CatalogHandler struct {
	pb.UnimplementedCatalogServiceServer
	svc *service.CatalogService
}

func NewCatalogHandler(svc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	product, err := h.svc.GetProduct(req.GetProductId())
	if err != nil {
		// Маппинг ошибок в правильные gRPC статусы
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "Товар с ID '%s' не найден", req.GetProductId())
		}
		if err.Error() == "product_id cannot be empty" {
			return nil, status.Error(codes.InvalidArgument, "ID товара не может быть пустым")
		}
		return nil, status.Errorf(codes.Internal, "Внутренняя ошибка сервера: %v", err)
	}

	// Конвертируем внутреннюю модель в Protobuf-сообщение
	return &pb.ProductResponse{
		Product: &pb.Product{
			Id:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			PriceCents:    product.PriceCents,
			StockQuantity: product.StockQuantity,
			Brand:         product.Brand,
		},
	}, nil
}

func (h *CatalogHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := h.svc.ListProducts()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Не удалось получить список товаров")
	}

	var pbProducts []*pb.Product
	for _, p := range products {
		pbProducts = append(pbProducts, &pb.Product{
			Id:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			PriceCents:    p.PriceCents,
			StockQuantity: p.StockQuantity,
			Brand:         p.Brand,
		})
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
		// next_page_token пока оставляем пустым, так как пагинацию добавим позже
	}, nil
}
