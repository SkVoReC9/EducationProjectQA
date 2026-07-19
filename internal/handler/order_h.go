package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Убедись, что путь соответствует твоему go.mod и структуре gen/
	pb "awesomeProject/gen/store/api/order/v1"
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

// Вспомогательный метод для маппинга внутренней модели заказа в gRPC Protobuf
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

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}

	order, err := h.svc.CreateOrder(req.GetUserId())
	if err != nil {
		if strings.Contains(err.Error(), "корзина пуста") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка создания заказа: %v", err)
	}

	return &pb.OrderResponse{
		Order: mapOrderToProto(order),
	}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id обязателен")
	}

	order, err := h.svc.GetOrder(req.GetOrderId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "заказ не найден")
	}

	return &pb.OrderResponse{
		Order: mapOrderToProto(order),
	}, nil
}
