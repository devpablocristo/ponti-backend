package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign"
)

func ProvideCampaignRepository(repo gorm.Repository) (campaign.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return campaign.NewRepository(repo), nil
}

func ProvideCampaignUseCases(repo campaign.Repository) campaign.UseCases {
	return campaign.NewUseCases(repo)
}

func ProvideCampaigHandler(server ginsrv.Server, usecases campaign.UseCases) *campaign.Handler {
	return campaign.NewHandler(server, usecases)
}
