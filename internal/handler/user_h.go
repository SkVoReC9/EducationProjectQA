package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "awesomeProject/gen/store/api/user/v1"
	"awesomeProject/internal/auth"
	"awesomeProject/internal/repository"
	"awesomeProject/internal/service"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func mapUserToProto(user repository.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: timestamppb.New(user.CreatedAt),
		Role:      user.Role,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email и password обязательны")
	}

	result, err := h.svc.Register(req.GetEmail(), req.GetPassword(), req.GetName())
	if err != nil {
		if err.Error() == "email и password обязательны" {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка регистрации: %v", err)
	}

	return &pb.AuthResponse{
		User:        mapUserToProto(result.User),
		AccessToken: result.AccessToken,
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email и password обязательны")
	}

	result, err := h.svc.Login(req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if err.Error() == "email и password обязательны" {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка входа: %v", err)
	}

	return &pb.AuthResponse{
		User:        mapUserToProto(result.User),
		AccessToken: result.AccessToken,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}

	user, err := h.svc.GetUser(req.GetUserId())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "пользователь не найден")
		}
		return nil, status.Errorf(codes.Internal, "ошибка получения пользователя: %v", err)
	}

	return &pb.UserResponse{User: mapUserToProto(user)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id обязателен")
	}
	if err := ensureCallerMatchesUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	if err := h.svc.DeleteUser(req.GetUserId()); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "пользователь не найден")
		}
		if errors.Is(err, repository.ErrUserHasActiveOrders) {
			return nil, status.Error(codes.FailedPrecondition, "нельзя удалить пользователя с активными заказами")
		}
		if err.Error() == "user_id обязателен" {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "ошибка удаления пользователя: %v", err)
	}
	return &pb.DeleteUserResponse{}, nil
}

func ensureCallerMatchesUser(ctx context.Context, userID string) error {
	caller := auth.UserIDFromContext(ctx)
	if caller == "" {
		return status.Error(codes.Unauthenticated, "требуется авторизация")
	}
	if caller != userID {
		return status.Error(codes.PermissionDenied, "user_id не совпадает с токеном")
	}
	return nil
}
