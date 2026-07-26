package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "awesomeProject/gen/store/api/cart/v1"
	"awesomeProject/internal/repository"
	"awesomeProject/internal/service"
)

type CartHandler struct {
	pb.UnimplementedCartServiceServer
	svc *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) buildCartResponseFromTotals(totals service.CartTotals) *pb.CartResponse {
	var items []*pb.CartItem
	for _, item := range totals.Cart.Items {
		items = append(items, &pb.CartItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	expires := ""
	if !totals.ExpiresAt.IsZero() {
		expires = totals.ExpiresAt.UTC().Format(time.RFC3339)
	}

	return &pb.CartResponse{
		Items:                items,
		TotalPriceCents:      totals.TotalPriceCents,
		SubtotalCents:        totals.SubtotalCents,
		DiscountCents:        totals.DiscountCents,
		AppliedPromocode:     totals.AppliedPromocode,
		ComboDiscountApplied: totals.ComboDiscountApplied,
		ExpiresAt:            expires,
	}
}

func (h *CartHandler) buildCartResponse(userID string) (*pb.CartResponse, error) {
	totals, err := h.svc.GetCart(userID)
	if err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
		return nil, status.Errorf(codes.Internal, "ошибка при получении корзины: %v", err)
	}
	return h.buildCartResponseFromTotals(totals), nil
}

func (h *CartHandler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	err := h.svc.AddItem(req.GetUserId(), req.GetProductId(), req.GetQuantity())
	if err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
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
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	err := h.svc.RemoveItem(req.GetUserId(), req.GetProductId())
	if err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return h.buildCartResponse(req.GetUserId())
}

func (h *CartHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}
	return h.buildCartResponse(req.GetUserId())
}

func (h *CartHandler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}
	if err := h.svc.ClearCart(req.GetUserId()); err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
		return nil, status.Errorf(codes.Internal, "ошибка очистки корзины: %v", err)
	}
	return h.buildCartResponse(req.GetUserId())
}

func (h *CartHandler) ApplyPromocode(ctx context.Context, req *pb.ApplyPromocodeRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	totals, err := h.svc.ApplyPromocode(req.GetUserId(), req.GetCode())
	if err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
		if strings.Contains(err.Error(), "промокод не найден") || errors.Is(err, repository.ErrPromocodeNotFound) {
			return nil, status.Error(codes.NotFound, "промокод не найден")
		}
		if strings.Contains(err.Error(), "promocode обязателен") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка применения промокода: %v", err)
	}
	return h.buildCartResponseFromTotals(totals), nil
}

func (h *CartHandler) ClearPromocode(ctx context.Context, req *pb.ClearPromocodeRequest) (*pb.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	totals, err := h.svc.ClearPromocode(req.GetUserId())
	if err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
		return nil, status.Errorf(codes.Internal, "ошибка удаления промокода: %v", err)
	}
	return h.buildCartResponseFromTotals(totals), nil
}

func mapUserNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrUserNotFound) || strings.Contains(err.Error(), "пользователь не найден") {
		return status.Error(codes.NotFound, "пользователь не найден")
	}
	return nil
}
