package client

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/Aruzhan38/smart-campus-generated/proto/user"
	"google.golang.org/grpc"
)

type UserClient interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type userClient struct {
	client pb.UserServiceClient
}

type User struct {
	ID    int
	Email string
}

func NewUserClient(conn *grpc.ClientConn) UserClient {
	return &userClient{client: pb.NewUserServiceClient(conn)}
}

func (uc *userClient) GetUserByID(ctx context.Context, id string) (*User, error) {
	resp, err := uc.client.GetUserById(ctx, &pb.GetUserByIdRequest{UserId: id})
	if err != nil {
		return nil, err
	}
	uid, err := strconv.Atoi(resp.User.Id)
	if err != nil {
		return nil, err
	}
	return &User{ID: uid, Email: resp.User.Email}, nil
}

func DialUserService(addr string) (*grpc.ClientConn, error) {
	if addr == "" {
		return nil, fmt.Errorf("user service address is empty")
	}
	return grpc.Dial(addr, grpc.WithInsecure())
}
