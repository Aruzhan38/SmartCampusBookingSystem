package grpc

import (
	"context"
	"errors"
	"room-service/internal/domain"
	"room-service/internal/usecase"
	"strconv"

	roompb "github.com/Aruzhan38/smart-campus-generated/proto/room"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type RoomServer struct {
	roompb.UnimplementedRoomServiceServer
	usecase usecase.RoomUsecase
}

func NewRoomServer(uc usecase.RoomUsecase) *RoomServer {
	return &RoomServer{usecase: uc}
}

func (s *RoomServer) CreateRoom(ctx context.Context, req *roompb.CreateRoomRequest) (*roompb.RoomResponse, error) {
	buildingID, err := parsePositiveUint(req.BuildingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "building_id must be a positive integer")
	}

	room := &domain.Room{
		BuildingID:  uint(buildingID),
		Number:      req.RoomNumber,
		Capacity:    req.Capacity,
		Description: req.RoomType,
	}

	if err := s.usecase.CreateRoom(ctx, room); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &roompb.RoomResponse{
		Room:    toProtoRoom(room),
		Message: "Room created successfully",
	}, nil
}

func (s *RoomServer) GetRoomById(ctx context.Context, req *roompb.GetRoomByIdRequest) (*roompb.RoomResponse, error) {
	id64, err := parsePositiveUint(req.RoomId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "room_id must be a positive integer")
	}

	room, err := s.usecase.GetRoomByID(ctx, uint(id64))
	if err != nil {
		return nil, roomServiceError(err)
	}

	return &roompb.RoomResponse{
		Room:    toProtoRoom(room),
		Message: "Room found",
	}, nil
}

func (s *RoomServer) ListRooms(ctx context.Context, req *roompb.ListRoomsRequest) (*roompb.ListRoomsResponse, error) {
	rooms, err := s.usecase.ListRooms(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &roompb.ListRoomsResponse{
		Rooms:   make([]*roompb.Room, 0, len(rooms)),
		Message: "Rooms loaded",
	}

	for _, room := range rooms {
		r := room
		resp.Rooms = append(resp.Rooms, toProtoRoom(&r))
	}

	return resp, nil
}

func (s *RoomServer) UpdateRoom(ctx context.Context, req *roompb.UpdateRoomRequest) (*roompb.RoomResponse, error) {
	roomID, err := parsePositiveUint(req.RoomId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "room_id must be a positive integer")
	}

	buildingID, err := parsePositiveUint(req.BuildingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "building_id must be a positive integer")
	}

	room := &domain.Room{
		ID:          uint(roomID),
		BuildingID:  uint(buildingID),
		Number:      req.RoomNumber,
		Capacity:    req.Capacity,
		Description: req.RoomType,
	}

	if err := s.usecase.UpdateRoom(ctx, room); err != nil {
		return nil, roomServiceError(err)
	}

	updatedRoom, err := s.usecase.GetRoomByID(ctx, room.ID)
	if err != nil {
		return nil, roomServiceError(err)
	}

	return &roompb.RoomResponse{
		Room:    toProtoRoom(updatedRoom),
		Message: "Room updated successfully",
	}, nil
}

func (s *RoomServer) SearchAvailableRooms(ctx context.Context, req *roompb.SearchAvailableRoomsRequest) (*roompb.ListRoomsResponse, error) {
	if req.MinCapacity < 1 {
		return nil, status.Error(codes.InvalidArgument, "min_capacity must be a positive integer")
	}

	rooms, err := s.usecase.SearchRoomsByCapacity(ctx, req.MinCapacity)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &roompb.ListRoomsResponse{
		Rooms:   make([]*roompb.Room, 0, len(rooms)),
		Message: "Rooms found",
	}

	for _, room := range rooms {
		r := room
		resp.Rooms = append(resp.Rooms, toProtoRoom(&r))
	}

	return resp, nil
}

func toProtoRoom(room *domain.Room) *roompb.Room {
	return &roompb.Room{
		Id:           strconv.Itoa(int(room.ID)),
		BuildingId:   strconv.Itoa(int(room.BuildingID)),
		BuildingName: "",
		RoomNumber:   room.Number,
		Capacity:     room.Capacity,
		RoomType:     room.Description,
		Status:       "Available",
		Equipment:    []string{},
		CreatedAt:    room.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func roomServiceError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "room not found")
	}
	return status.Error(codes.Internal, err.Error())
}

func parsePositiveUint(value string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}
