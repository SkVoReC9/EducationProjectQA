package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Замени "store" на имя твоего модуля
	pb "awesomeProject/gen/store/api/cart/v1"
	"awesomeProject/internal/service"
)

type CartHandler struct {
	pb.UnimplementedCartServiceServer
	svc *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

// Вспомогательный метод для формирования ответа
func (h *CartHandler) buildCartResponse(userID string) (*pb.CartResponse, error) {
	cart, totalPrice, err := h.svc.GetCart(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка при получении корзины: %v", err)
	}

	var items []*pb.CartItem
	for _, item := range cart.Items {
		items = append(items, &pb.CartItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &pb.CartResponse{
		Items:           items,
		TotalPriceCents: totalPrice,
	}, nil
}

func (h *CartHandler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}

	err := h.svc.AddItem(req.GetUserId(), req.GetProductId(), req.GetQuantity())
	if err != nil {
		// Маппинг ошибок бизнес-логики в gRPC статусы
		if strings.Contains(err.Error(), "quantity must be greater than zero") || strings.Contains(err.Error(), "max 99") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if strings.Contains(err.Error(), "product validation failed") {
			return nil, status.Error(codes.NotFound, "товар не найден в каталоге")
		}
		return nil, status.Errorf(codes.Internal, "внутренняя ошибка: %v", err)
	}

	return h.buildCartResponse(req.GetUserId())
}

func (h *CartHandler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}

	err := h.svc.RemoveItem(req.GetUserId(), req.GetProductId())
	if err != nil {
		// Здесь выстрелит наш БАГ №3 (CRITICAL: pointer to nil item...)
		// Мы намеренно отдаем его как Internal ошибку, чтобы напугать тестировщика
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return h.buildCartResponse(req.GetUserId())
}

func (h *CartHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	return h.buildCartResponse(req.GetUserId())
}

func (h *CartHandler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	_ = h.svc.ClearCart(req.GetUserId())
	return h.buildCartResponse(req.GetUserId())
}
