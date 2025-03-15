package authe

import (
	"context"
	"errors"
	"fmt"
	"time"

	types "github.com/devpablocristo/monorepo/pkg/types"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/authe/usecases/domain"
	support "github.com/devpablocristo/monorepo/projects/qh/internal/authe/usecases/support"
)

type useCases struct {
	cache      Cache
	jwtService JwtService
	httpClient HttpClient
}

func NewUseCases(
	ch Cache,
	js JwtService,
	hc HttpClient,

) UseCases {
	return &useCases{
		cache:      ch,
		jwtService: js,
		httpClient: hc,
	}
}

func (u *useCases) JwtLogin(ctx context.Context, username, email, password string) (*domain.Token, error) {
	return nil, nil
}

func (u *useCases) GenerateLinkTokens(ctx context.Context, userID string) (*domain.Token, error) {
	if userID == "" {
		return nil, errors.New("userID is empty")
	}

	token, err := u.jwtService.GenerateHrTokens(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	err = u.cache.StoreToken(ctx, userID, token)
	if err != nil {
		return nil, fmt.Errorf("failed storing refresh token: %w", err)
	}

	return token, nil
}

func (u *useCases) PepLogin(ctx context.Context, username, email, password string) (*domain.Token, error) {
	nameCred, passCred, err := support.GetCredentials(username, email, password)
	if err != nil {
		// Retornamos un error de dominio al no poder obtener credenciales
		return nil, types.NewError(types.ErrInvalidInput, "failed to get credentials", err)
	}

	// Intentar recuperar el token desde caché
	token, err := u.cache.RetrieveToken(ctx, nameCred)

	switch {
	case err == nil:
		// Caso 1: No hubo error al buscar en caché
		if token != nil && token.AccessExpiresAt.After(time.Now()) {
			// El token aún es válido
			fmt.Println("Token is still valid until:", token.AccessExpiresAt)
			return token, nil
		}
		// Si está expirado, continúa para obtener uno nuevo

	case types.IsTokenNotFoundError(err):
		// Caso 2: Token no existe en caché
		fmt.Println("Token not found in cache, requesting a new one")

	default:
		// Caso 3: Error inesperado al recuperar el token
		return nil, types.NewError(types.ErrOperationFailed, "failed to retrieve token from cache", err)
	}

	// Obtener el token externo desde la API de PEP
	externalToken, err := u.httpClient.GetAccessTokenPep(ctx, nameCred, passCred)
	if err != nil {
		// No pudimos obtener el token desde la API de Pep
		return nil, types.NewError(types.ErrOperationFailed, "failed to get pep access token", err)
	}

	// Extraer las claims del token externo
	extractedClaims, err := u.jwtService.ExtractClaimsFromExternalToken(externalToken.AccessToken)
	if err != nil {
		return nil, types.NewError(types.ErrInvalidInput, "failed to extract claims from external token", err)
	}

	// Obtener el userID y email desde las claims
	userID, ok := extractedClaims["sub"].(string)
	if !ok || userID == "" {
		return nil, types.NewError(types.ErrInvalidInput, "invalid or missing userID in token claims", nil)
	}

	// Generar token interno (nuestro JWT)
	token, err = u.jwtService.GenerateHrTokens(ctx, userID)
	if err != nil {
		return nil, types.NewError(types.ErrOperationFailed, "failed to generate internal token", err)
	}

	// Almacenar el nuevo token en la caché
	if err = u.cache.StoreToken(ctx, nameCred, token); err != nil {
		return nil, types.NewError(types.ErrOperationFailed, "failed storing token in cache", err)
	}

	return token, nil
}

func (u *useCases) Auth0Login(ctx context.Context, username, email, password string) (*domain.Token, error) {
	return nil, nil
}
