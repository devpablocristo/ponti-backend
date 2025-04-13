package authe

import (
	"context"
	"fmt"

	resty "github.com/alphacodinggroup/euxcel-backend/pkg/http/clients/resty"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe/usecases/domain"
	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/config"
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
