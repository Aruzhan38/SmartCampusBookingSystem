package client

import (
	"context"

	roompb "github.com/Aruzhan38/smart-campus-generated/proto/room"
	"google.golang.org/grpc"
)

type RoomClient interface {
	GetRooms(ctx context.Context) (*roompb.ListRoomsResponse, error)
	GetRoomByID(ctx context.Context, id string) (*roompb.RoomResponse, error)
	CreateRoom(
		ctx context.Context,
		roomNumber string,
		capacity int32,
		buildingID string,
		description string,
	) (*roompb.RoomResponse, error)
	UpdateRoom(
		ctx context.Context,
		roomID string,
		roomNumber string,
		capacity int32,
		buildingID string,
		description string,
	) (*roompb.RoomResponse, error)
	SearchRoomsByCapacity(ctx context.Context, minCapacity int32) (*roompb.ListRoomsResponse, error)
}

type roomClient struct {
	client roompb.RoomServiceClient
}

func NewRoomClient(conn *grpc.ClientConn) RoomClient {
	return &roomClient{
		client: roompb.NewRoomServiceClient(conn),
	}
}

func (c *roomClient) GetRooms(ctx context.Context) (*roompb.ListRoomsResponse, error) {
	return c.client.ListRooms(ctx, &roompb.ListRoomsRequest{})
}

func (c *roomClient) GetRoomByID(ctx context.Context, id string) (*roompb.RoomResponse, error) {
	return c.client.GetRoomById(ctx, &roompb.GetRoomByIdRequest{
		RoomId: id,
	})
}

func (c *roomClient) CreateRoom(
	ctx context.Context,
	roomNumber string,
	capacity int32,
	buildingID string,
	description string,
) (*roompb.RoomResponse, error) {

	return c.client.CreateRoom(ctx, &roompb.CreateRoomRequest{
		RoomNumber: roomNumber,
		Capacity:   capacity,
		BuildingId: buildingID,
		RoomType:   description,
	})
}

func (c *roomClient) UpdateRoom(
	ctx context.Context,
	roomID string,
	roomNumber string,
	capacity int32,
	buildingID string,
	description string,
) (*roompb.RoomResponse, error) {
	return c.client.UpdateRoom(ctx, &roompb.UpdateRoomRequest{
		RoomId:     roomID,
		RoomNumber: roomNumber,
		Capacity:   capacity,
		BuildingId: buildingID,
		RoomType:   description,
	})
}

func (c *roomClient) SearchRoomsByCapacity(ctx context.Context, minCapacity int32) (*roompb.ListRoomsResponse, error) {
	return c.client.SearchAvailableRooms(ctx, &roompb.SearchAvailableRoomsRequest{
		MinCapacity: minCapacity,
	})
}
