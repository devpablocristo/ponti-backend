package authe

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	client "github.com/devpablocristo/monorepo/pkg/microservices/go-micro/v4/grpc-client"
// 	domain "github.com/devpablocristo/monorepo/projects/qh/internal/authe/usecases/domain"
// 	"github.com/devpablocristo/monorepo/projects/qh/internal/user/grpc/pb"
// )

// type grpcClient struct {
// 	client     client.Client
// 	serverName string
// }

// // NewGrpcClient crea un nuevo cliente gRPC para interactuar con el servicio de usuarios
// func NewGrpcClient() (GrpcClient, error) {
// 	c, err := client.Bootstrap() // Usamos tu método Bootstrap del SDK go-micro
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize Go Micro gRPC client: %w", err)
// 	}

// 	return &grpcClient{
// 		client: c,
// 	}, nil
// }

// func (g *grpcClient) GetClient() client.Client {
// 	return g.client
// }

// func (g *grpcClient) GetUserUUID(ctx context.Context, cred *domain.LoginCredentials) (string, error) {
// 	req := &pb.GetUserRequest{
// 		Username:     cred.Username,
// 		PasswordHash: cred.PasswordHash,
// 	}

// 	client := g.client.GetClient()
// 	request := client.NewRequest(g.client.GetServerName(), "GetUserUUID", req)

// 	var res pb.GetUserResponse

// 	// Crear un contexto con timeout para la llamada gRPC
// 	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	if err := g.client.GetClient().Call(callCtx, request, &res); err != nil {
// 		return "", fmt.Errorf("error calling GetUserUUID: %w", err)
// 	}

// 	return res.UUID, nil
// }
