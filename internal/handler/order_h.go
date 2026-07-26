package handler

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "awesomeProject/gen/store/api/order/v1"
	"awesomeProject/internal/auth"
	"awesomeProject/internal/repository"
	"awesomeProject/internal/service"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func mapOrderToProto(order repository.Order) *pb.Order {
	var items []*pb.OrderItem
	for _, item := range order.Items {
		items = append(items, &pb.OrderItem{
			ProductId:  item.ProductID,
			Quantity:   item.Quantity,
			PriceCents: item.PriceCents,
		})
	}

	return &pb.Order{
		Id:               order.ID,
		UserId:           order.UserID,
		Items:            items,
		TotalAmountCents: order.TotalAmountCents,
		Status:           pb.OrderStatus(order.Status),
	}
}

func callerIsAdmin(ctx context.Context) bool {
	return auth.RoleFromContext(ctx) == "admin"
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	order, err := h.svc.CreateOrder(req.GetUserId())
	if err != nil {
		if mapped := mapUserNotFound(err); mapped != nil {
			return nil, mapped
		}
		if strings.Contains(err.Error(), "корзина пуста") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка создания заказа: %v", err)
	}

	return &pb.OrderResponse{Order: mapOrderToProto(order)}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id обязателен")
	}

	order, err := h.svc.GetOrder(req.GetOrderId(), auth.UserIDFromContext(ctx), callerIsAdmin(ctx))
	if err != nil {
		return nil, mapOrderErr(err)
	}

	return &pb.OrderResponse{Order: mapOrderToProto(order)}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id обязателен")
	}

	order, err := h.svc.CancelOrder(req.GetOrderId(), auth.UserIDFromContext(ctx), callerIsAdmin(ctx))
	if err != nil {
		return nil, mapOrderErr(err)
	}
	return &pb.OrderResponse{Order: mapOrderToProto(order)}, nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id обязателен")
	}
	if req.GetFromStatus() == pb.OrderStatus_ORDER_STATUS_UNSPECIFIED || req.GetToStatus() == pb.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "from_status и to_status обязательны")
	}

	order, err := h.svc.UpdateOrderStatus(
		req.GetOrderId(),
		int32(req.GetFromStatus()),
		int32(req.GetToStatus()),
		auth.UserIDFromContext(ctx),
		callerIsAdmin(ctx),
	)
	if err != nil {
		return nil, mapOrderErr(err)
	}
	return &pb.OrderResponse{Order: mapOrderToProto(order)}, nil
}

func mapOrderErr(err error) error {
	if errors.Is(err, repository.ErrOrderNotFound) || errors.Is(err, service.ErrOrderNotFound) {
		return status.Error(codes.NotFound, "заказ не найден")
	}
	if errors.Is(err, service.ErrPermissionDenied) {
		return status.Error(codes.PermissionDenied, "нет доступа к заказу")
	}
	if errors.Is(err, service.ErrInvalidTransition) {
		return status.Error(codes.FailedPrecondition, "недопустимый переход статуса")
	}
	if errors.Is(err, service.ErrStatusMismatch) {
		return status.Error(codes.FailedPrecondition, "текущий статус не совпадает с from_status")
	}
	if strings.Contains(err.Error(), "order_id не может быть пустым") {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Errorf(codes.Internal, "ошибка заказа: %v", err)
}
