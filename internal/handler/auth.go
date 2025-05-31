package handler

import (
	"context"

	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/service"
	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/utils"
	"github.com/berrylradianh/go-grpc-ecommerce-be/pb/auth"
)

type authHandler struct {
	auth.UnimplementedAuthServiceServer
	authService service.IAuthService
}

func (ah *authHandler) Register(ctx context.Context, request *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	validationErrors, err := utils.CheckValidation(request)
	if err != nil {
		return nil, err
	}
	if validationErrors != nil {
		return &auth.RegisterResponse{
			Base: utils.ValidationErrorResponse(validationErrors),
		}, nil
	}

	res, err := ah.authService.Register(ctx, request)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ah *authHandler) Logout(ctx context.Context, request *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	validationErrors, err := utils.CheckValidation(request)
	if err != nil {
		return nil, err
	}
	if validationErrors != nil {
		return &auth.LogoutResponse{
			Base: utils.ValidationErrorResponse(validationErrors),
		}, nil
	}

	res, err := ah.authService.Logout(ctx, request)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ah *authHandler) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	validationErrors, err := utils.CheckValidation(request)
	if err != nil {
		return nil, err
	}
	if validationErrors != nil {
		return &auth.LoginResponse{
			Base: utils.ValidationErrorResponse(validationErrors),
		}, nil
	}

	res, err := ah.authService.Login(ctx, request)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewAuthHandler(authService service.IAuthService) *authHandler {
	return &authHandler{
		authService: authService,
	}
}
