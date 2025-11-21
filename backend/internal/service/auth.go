package service

import (
	"context"

	"github.com/amityadav/landr/internal/core"
	"github.com/amityadav/landr/pkg/pb/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	core *core.AuthCore
}

func NewAuthService(c *core.AuthCore) *AuthService {
	return &AuthService{
		core: c,
	}
}

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, sessionToken, err := s.core.LoginUser(ctx, req.GoogleIdToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "login failed: %v", err)
	}

	return &auth.LoginResponse{
		SessionToken: sessionToken,
		User:         user,
	}, nil
}
