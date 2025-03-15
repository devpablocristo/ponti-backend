package authe

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	// Ajusta estos imports a las rutas reales de tu proyecto:
// 	pb "github.com/devpablocristo/sdk/pb"
// 	sdkgrpc "github.com/devpablocristo/sdk/pkg/grpc/server"
// 	defs "github.com/devpablocristo/sdk/pkg/grpc/server/defs"
// 	ports "github.com/devpablocristo/sdk/sg/auth/internal/core/ports"
// )

// // GrpcServer implementa la interfaz ports.GrpcServer.
// // Además, "embebe" pb.UnimplementedAuthServiceServer para cumplir con el contrato de gRPC.
// type GrpcServer struct {
// 	pb.UnimplementedAuthServiceServer
// 	ucs    ports.UseCases
// 	server defs.Server
// }

// // NewGrpcServer inicializa una nueva instancia de GrpcServer y registra el AuthService.
// func NewGrpcServer(u ports.UseCases) (ports.GrpcServer, error) {
// 	// Inicializar el servidor gRPC usando tu paquete sdkgrpc.
// 	s, err := sdkgrpc.Bootstrap()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to bootstrap gRPC server: %w", err)
// 	}

// 	grpcServer := &GrpcServer{
// 		server: s,
// 		ucs:    u,
// 	}

// 	// Registrar el servicio AuthService usando RegisterService.
// 	// Ajusta el primer parámetro (context) si tu librería lo requiere o no.
// 	grpcServer.server.RegisterService(context.Background(), &pb.AuthService_ServiceDesc, grpcServer)

// 	return grpcServer, nil
// }

// // GenerateTokens es un método de la interfaz AuthService (definido en tu .proto).
// func (g *GrpcServer) GenerateTokens(ctx context.Context, req *pb.VerificationTokenRequest) (*pb.VerificationTokenResponse, error) {
// 	accessToken, refreshToken, err := g.ucs.GenerateTokens(ctx, req.GetUserUuid())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to generate verification tokens: %w", err)
// 	}

// 	return &pb.VerificationTokenResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken,
// 	}, nil
// }

// // Start inicia el servidor gRPC.
// func (g *GrpcServer) Start(ctx context.Context) error {
// 	log.Println("Starting gRPC server...")
// 	return g.server.Start(ctx)
// }

// // Stop detiene el servidor gRPC.
// func (g *GrpcServer) Stop() error {
// 	log.Println("Stopping gRPC server...")
// 	return g.server.Stop()
// }
