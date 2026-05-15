package grpc

import (
	"context"
	"strconv"
	"user-service/internal/usecase"

	pb "github.com/Aruzhan38/smart-campus-generated/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userServer struct {
	pb.UnimplementedUserServiceServer
	usecase usecase.UserUsecase
}

func NewUserServer(usecase usecase.UserUsecase) pb.UserServiceServer {
	return &userServer{usecase: usecase}
}

func (s *userServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	err := s.usecase.RegisterUser(req.FullName, req.Email, req.Password, req.Role)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterUserResponse{Message: "User registered successfully"}, nil
}

func (s *userServer) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	token, err := s.usecase.LoginUser(req.Email, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
		}
		return nil, err
	}
	return &pb.LoginUserResponse{Token: token}, nil
}

func (s *userServer) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	id, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, err
	}
	user, err := s.usecase.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserByIdResponse{
		User: &pb.User{
			Id:       strconv.Itoa(user.ID),
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *userServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	user, err := s.usecase.ValidateToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false, Message: err.Error()}, nil
	}
	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: strconv.Itoa(user.ID),
		Role:   user.Role,
	}, nil
}
