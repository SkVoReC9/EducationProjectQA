package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "awesomeProject/gen/store/api/promo/v1"
	"awesomeProject/internal/repository"
	"awesomeProject/internal/service"
)

type PromoHandler struct {
	pb.UnimplementedPromoServiceServer
	svc *service.PromoService
}

func NewPromoHandler(svc *service.PromoService) *PromoHandler {
	return &PromoHandler{svc: svc}
}

func mapPromoToProto(p repository.Promocode) *pb.Promocode {
	out := &pb.Promocode{
		Code:          p.Code,
		DiscountValue: p.DiscountValue,
		Active:        p.Active,
	}
	switch p.DiscountType {
	case repository.DiscountPercent:
		out.DiscountType = pb.DiscountType_DISCOUNT_TYPE_PERCENT
	case repository.DiscountFixedCents:
		out.DiscountType = pb.DiscountType_DISCOUNT_TYPE_FIXED_CENTS
	}
	if p.ExpiresAt != nil {
		out.ExpiresAt = p.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return out
}

func discountTypeFromProto(t pb.DiscountType) (string, error) {
	switch t {
	case pb.DiscountType_DISCOUNT_TYPE_PERCENT:
		return repository.DiscountPercent, nil
	case pb.DiscountType_DISCOUNT_TYPE_FIXED_CENTS:
		return repository.DiscountFixedCents, nil
	default:
		return "", errors.New("discount_type обязателен")
	}
}

func (h *PromoHandler) CreatePromocode(ctx context.Context, req *pb.CreatePromocodeRequest) (*pb.PromocodeResponse, error) {
	dtype, err := discountTypeFromProto(req.GetDiscountType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	expires, err := service.ParseExpiresAt(req.GetExpiresAt())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// New codes are always created active; deactivate via UpdatePromocode.
	saved, err := h.svc.Create(req.GetCode(), dtype, req.GetDiscountValue(), true, expires)
	if err != nil {
		if strings.Contains(err.Error(), "обязателен") || strings.Contains(err.Error(), "должен") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка создания промокода: %v", err)
	}
	return &pb.PromocodeResponse{Promocode: mapPromoToProto(saved)}, nil
}

func (h *PromoHandler) ListPromocodes(ctx context.Context, req *pb.ListPromocodesRequest) (*pb.ListPromocodesResponse, error) {
	list, err := h.svc.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка списка промокодов: %v", err)
	}
	out := make([]*pb.Promocode, 0, len(list))
	for _, p := range list {
		out = append(out, mapPromoToProto(p))
	}
	return &pb.ListPromocodesResponse{Promocodes: out}, nil
}

func (h *PromoHandler) UpdatePromocode(ctx context.Context, req *pb.UpdatePromocodeRequest) (*pb.PromocodeResponse, error) {
	dtype, err := discountTypeFromProto(req.GetDiscountType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	expires, err := service.ParseExpiresAt(req.GetExpiresAt())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	saved, err := h.svc.Update(req.GetCode(), dtype, req.GetDiscountValue(), req.GetActive(), expires)
	if err != nil {
		if errors.Is(err, repository.ErrPromocodeNotFound) {
			return nil, status.Error(codes.NotFound, "промокод не найден")
		}
		if strings.Contains(err.Error(), "обязателен") || strings.Contains(err.Error(), "должен") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка обновления промокода: %v", err)
	}
	return &pb.PromocodeResponse{Promocode: mapPromoToProto(saved)}, nil
}

func (h *PromoHandler) DeletePromocode(ctx context.Context, req *pb.DeletePromocodeRequest) (*pb.DeletePromocodeResponse, error) {
	if err := h.svc.Delete(req.GetCode()); err != nil {
		if errors.Is(err, repository.ErrPromocodeNotFound) {
			return nil, status.Error(codes.NotFound, "промокод не найден")
		}
		if strings.Contains(err.Error(), "обязателен") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка удаления промокода: %v", err)
	}
	return &pb.DeletePromocodeResponse{}, nil
}
