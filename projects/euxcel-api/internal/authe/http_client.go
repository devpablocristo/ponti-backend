package authe

import (
	"context"
	"errors"
	"fmt"

	resty "github.com/alphacodinggroup/euxcel-backend/pkg/http/clients/resty"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/authe/http-client/dto"
	domain "github.com/alphacodinggroup/euxcel-backend/internal/authe/usecases/domain"
	config "github.com/alphacodinggroup/euxcel-backend/internal/config"
)

type httpClient struct {
	client resty.Client
	config config.Loader
}

func NewHttpClient(client resty.Client, config config.Loader) HttpClient {
	return &httpClient{
		client: client,
		config: config,
	}
}

func (a *httpClient) GetAccessToken(ctx context.Context, endpoint string, payload any) (*domain.Token, error) {
	var token *domain.Token
	req := a.client.GetClient().R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&token)
	resp, err := req.Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error executing POST: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("request error, status code: %d", resp.StatusCode())
	}
	return token, nil
}

func (a *httpClient) GetAccessTokenPep(ctx context.Context, username, password string) (*domain.Token, error) {

	// Obtener el endpoint de login de PEP desde la configuración
	loginEndpoint := a.config.GetPepConfig().Endpoints.Login

	// Obtener el BaseURL desde la configuración
	baseURL := a.config.GetPepConfig().BaseURL

	// Construir la URL completa
	fullURL := fmt.Sprintf("%s%s", baseURL, loginEndpoint)

	formData := map[string]string{
		"grant_type":    "password",
		"username":      username,
		"password":      password,
		"scope":         "",
		"client_id":     "string",
		"client_secret": "string",
	}

	// var tokenResponse dto.AccessTokenResponse
	var token *dto.Token

	req := a.client.GetClient().R().
		SetContext(ctx).
		SetFormData(formData).
		SetResult(&token)

	resp, err := req.Post(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to execute POST request: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode())
	}

	err = token.ParseJwtClaims()
	if err != nil {
		return nil, errors.New("dto parsing failed")
	}

	fmt.Println(token.AccessExpiresAt)

	return token.ToDomain(), nil
}
