package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey int

const (
	userIDKey ctxKey = 1
	roleKey   ctxKey = 2
)

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func ContextWithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

func RoleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(roleKey).(string)
	return v
}

func UnaryInterceptor(jwt *Manager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !requiresAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		token, err := bearerFromMetadata(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "требуется Authorization: Bearer <token>")
		}

		userID, _, role, err := jwt.Parse(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "недействительный или просроченный токен")
		}

		ctx = ContextWithUserID(ctx, userID)
		ctx = ContextWithRole(ctx, role)

		if requiresAdmin(info.FullMethod) && role != "admin" {
			return nil, status.Error(codes.PermissionDenied, "требуется роль admin")
		}

		return handler(ctx, req)
	}
}

func requiresAuth(fullMethod string) bool {
	return strings.Contains(fullMethod, "CartService") ||
		strings.Contains(fullMethod, "OrderService") ||
		strings.Contains(fullMethod, "PromoService") ||
		strings.HasSuffix(fullMethod, "/DeleteUser")
}

func requiresAdmin(fullMethod string) bool {
	return strings.Contains(fullMethod, "PromoService")
}

func bearerFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	vals := md.Get("authorization")
	if len(vals) == 0 {
		vals = md.Get("Authorization")
	}
	if len(vals) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}

	parts := strings.SplitN(vals[0], " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}
	return parts[1], nil
}
