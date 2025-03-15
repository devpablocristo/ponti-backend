package rating

type useCases struct {
	grpcClient GrpcClient
}

// NewUseCases crea una nueva instancia de useCases
func NewUseCases(gc GrpcClient) UseCases {
	return &useCases{
		grpcClient: gc,
	}
}

// Login maneja la lógica de autenticación de usuario
// func (s *useCases) Login(ctx context.Context, creds *entities.LoginCredentials) (*entities.Token, error) {
// 	userID, err := s.grpcClient.GetUserID(ctx, creds)
// 	if err != nil {
// 		return nil, fmt.Errorf("error al obtener el ID del usuario: %w", err)
// 	}

// 	token, err := s.jwtService.GenerateToken(userID)
// 	if err != nil {
// 		return nil, fmt.Errorf("error generando el token de autenticación: %w", err)
// 	}

// 	// Devuelve el token generado
// 	return token, nil
// }
