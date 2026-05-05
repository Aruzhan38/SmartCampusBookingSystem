package client

import (
	"context"

	pb "github.com/Aruzhan38/smart-campus-generated/proto/room"

	"google.golang.org/grpc"
)

type RoomClient interface {
	GetRooms(ctx context.Context) (*pb.ListRoomsResponse, error)
}

type roomClient struct {
	client pb.RoomServiceClient
}

func NewRoomClient(conn *grpc.ClientConn) RoomClient {
	return &roomClient{client: pb.NewRoomServiceClient(conn)}
}

func (rc *roomClient) GetRooms(ctx context.Context) (*pb.ListRoomsResponse, error) {
	return rc.client.ListRooms(ctx, &pb.ListRoomsRequest{})
}
