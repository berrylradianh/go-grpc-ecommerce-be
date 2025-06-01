package grpcmiddleware

import (
	"context"

	jwtEntity "github.com/berrylradianh/go-grpc-ecommerce-be/internal/entity/jwt"
	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/utils"
	gocache "github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
)

type authMiddleware struct {
	cacheService *gocache.Cache
}

func (am *authMiddleware) Middleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if info.FullMethod == "/auth.AuthService/Login" || info.FullMethod == "/auth.AuthService/Register" {
		return handler(ctx, req)
	}

	tokenStr, err := jwtEntity.ParseTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	_, ok := am.cacheService.Get(tokenStr)
	if ok {
		return nil, utils.UnauthenticatedResponse()
	}

	claims, err := jwtEntity.GetClaimsFromToken(tokenStr)
	if err != nil {
		return nil, err
	}

	ctx = claims.SetToContext(ctx)

	res, err := handler(ctx, req)
	return res, err
}

func NewAuthMiddleware(cacheService *gocache.Cache) *authMiddleware {
	return &authMiddleware{
		cacheService: cacheService,
	}
}
