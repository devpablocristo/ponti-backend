package authe

import (
	"context"
	"fmt"
	"time"

	jwt0 "github.com/golang-jwt/jwt/v5"

	jwt "github.com/devpablocristo/monorepo/pkg/authe/jwt/v5"

	dto "github.com/devpablocristo/monorepo/projects/qh/internal/authe/jwt-service/dto"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/authe/usecases/domain"
	config "github.com/devpablocristo/monorepo/projects/qh/internal/config"
)

type jwtService struct {
	jwtService jwt.Service
	config     config.Loader
}

func NewJwtService(js jwt.Service, cf config.Loader) (JwtService, error) {
	return &jwtService{
		jwtService: js,
		config:     cf,
	}, nil
}

func (j *jwtService) GenerateHrTokens(ctx context.Context, userID string) (*domain.Token, error) {

	accessExp := j.config.GetHrConfig().AccessExpirationMinutes
	refreshExp := j.config.GetHrConfig().RefreshExpirationMinutes

	jwtToken, err := j.jwtService.GenerateTokens(ctx, userID, accessExp, refreshExp)
	if err != nil {
		return nil, fmt.Errorf("error trying to generate tokens: %w", err)
	}

	return dto.ToTokenDomain(jwtToken), nil
}

func (j *jwtService) GenerateLinkTokens(ctx context.Context, userID string) (*domain.Token, error) {
	accessExp := j.config.GetAssessmentConfig().AccessExpirationMinutes
	refreshExp := j.config.GetAssessmentConfig().RefreshExpirationMinutes

	jwtToken, err := j.jwtService.GenerateTokens(ctx, userID, accessExp, refreshExp)
	if err != nil {
		return nil, fmt.Errorf("error trying to generate tokens: %w", err)
	}

	return dto.ToTokenDomain(jwtToken), nil
}

func (j *jwtService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	jwtClaims, err := j.jwtService.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error trying to validate token: %w", err)
	}
	return dto.ToTokenClaimsDomain(jwtClaims), nil
}

func (j *jwtService) GetAccessExpiration(ctx context.Context) time.Duration {
	return j.jwtService.GetAccessExpiration()
}

func (j *jwtService) GetRefreshExpiration(ctx context.Context) time.Duration {
	return j.jwtService.GetRefreshExpiration()
}

func (j *jwtService) ExtractClaimsFromExternalToken(tokenString string) (map[string]any, error) {
	// Define las keys de las claims que deseas extraer (en este caso, "sub")
	claimKeys := []string{"sub"}

	// Creamos un parser y decodificamos el token sin verificar la firma.
	// Importante: al no verificar la firma, la validez del token no está garantizada.
	parser := new(jwt0.Parser)
	token, _, err := parser.ParseUnverified(tokenString, jwt0.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	// Convertimos las claims a jwt.MapClaims para acceder a sus valores.
	claims, ok := token.Claims.(jwt0.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Extraemos únicamente las claims que nos interesan.
	extractedClaims := make(map[string]any)
	for _, key := range claimKeys {
		if value, exists := claims[key]; exists {
			extractedClaims[key] = value
		}
	}

	return extractedClaims, nil
}
