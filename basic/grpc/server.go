package grpc

import (
	"context"
	"fmt"
)

type Server struct {
	UnimplementedUserServiceServer
}

var _ UserServiceServer = &Server{}

func (s Server) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	fmt.Println(request.GetId(), "i----d")
	return &GetByIdResponse{
		User: &User{
			Id:   123,
			Name: "abcd",
		},
	}, nil
}
