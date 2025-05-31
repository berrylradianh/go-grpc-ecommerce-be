package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/entity"
	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/repository"
	"github.com/berrylradianh/go-grpc-ecommerce-be/internal/utils"
	"github.com/berrylradianh/go-grpc-ecommerce-be/pb/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type IAuthService interface {
	Register(ctx context.Context, request *auth.RegisterRequest) (*auth.RegisterResponse, error)
	Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error)
	Logout(ctx context.Context, request *auth.LogoutRequest) (*auth.LogoutResponse, error)
}

type authService struct {
	authRepository repository.IAuthRepository
	cacheService   *gocache.Cache
}

func (as *authService) Register(ctx context.Context, request *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	if request.Password != request.PasswordConfirmation {
		return &auth.RegisterResponse{
			Base: utils.BadRequestResponse("Password and Password Confirmation does not match"),
		}, nil
	}

	user, err := as.authRepository.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return &auth.RegisterResponse{
			Base: utils.BadRequestResponse("Email already exist"),
		}, nil
	}

	// Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// insert to db
	newUser := &entity.User{
		ID:        uuid.NewString(),
		FullName:  request.FullName,
		Email:     request.Email,
		Password:  string(hashedPassword),
		RoleCode:  entity.UserRoleCustomer,
		CreatedAt: time.Now(),
		CreatedBy: &request.FullName,
	}

	err = as.authRepository.InsertUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &auth.RegisterResponse{
		Base: utils.SuccessResponse("User is registered successfully"),
	}, nil
}

func (as *authService) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, err := as.authRepository.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &auth.LoginResponse{
			Base: utils.BadRequestResponse("Email or Password is incorrect"),
		}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
		}
		return nil, err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.RoleCode,
	})
	secretKey := os.Getenv("JWT_SECRET_KEY")
	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		Base:        utils.SuccessResponse("User is logged in successfully"),
		AccessToken: accessToken,
	}, nil
}

func (as *authService) Logout(ctx context.Context, request *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	bearerToken, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	if len(bearerToken) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	tokenSplit := strings.Split(bearerToken[0], " ")
	if len(tokenSplit) != 2 {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	if tokenSplit[0] != "Bearer" {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	jwtToken := tokenSplit[1]
	tokenClaims, err := jwt.ParseWithClaims(jwtToken, &entity.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	if !tokenClaims.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	var claims *entity.JwtClaims
	if claims, ok = tokenClaims.Claims.(*entity.JwtClaims); !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}
	as.cacheService.Set(jwtToken, "", time.Duration(claims.ExpiresAt.Time.Unix()-time.Now().Unix())*time.Second)

	return &auth.LogoutResponse{
		Base: utils.SuccessResponse("User is logged out successfully"),
	}, nil
}

func NewAuthService(authRepository repository.IAuthRepository, cacheService *gocache.Cache) IAuthService {
	return &authService{
		authRepository: authRepository,
		cacheService:   cacheService,
	}
}
