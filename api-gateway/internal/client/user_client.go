package client

import (
	"api-gateway/internal/domain"
	"context"
	"errors"
	"strconv"

	pb "github.com/Aruzhan38/smart-campus-generated/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserClient interface {
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
	LoginUser(ctx context.Context, email, password string) (string, error)
	RegisterUser(ctx context.Context, fullName, email, password, role string) error
}

type userClient struct {
	client pb.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) UserClient {
	return &userClient{client: pb.NewUserServiceClient(conn)}
}

func (uc *userClient) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	resp, err := uc.client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	if !resp.Valid {
		return nil, errors.New(resp.Message)
	}
	// Get full user info
	userResp, err := uc.client.GetUserById(ctx, &pb.GetUserByIdRequest{UserId: resp.UserId})
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(userResp.User.Id)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:       id,
		FullName: userResp.User.FullName,
		Email:    userResp.User.Email,
		Role:     userResp.User.Role,
	}, nil
}

func (uc *userClient) LoginUser(ctx context.Context, email, password string) (string, error) {
	resp, err := uc.client.LoginUser(ctx, &pb.LoginUserRequest{Email: email, Password: password})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			return "", errors.New("Invalid email or password")
		}
		return "", err
	}
	return resp.Token, nil
}

func (uc *userClient) RegisterUser(ctx context.Context, fullName, email, password, role string) error {
	_, err := uc.client.RegisterUser(ctx, &pb.RegisterUserRequest{FullName: fullName, Email: email, Password: password, Role: role})
	return err
}
