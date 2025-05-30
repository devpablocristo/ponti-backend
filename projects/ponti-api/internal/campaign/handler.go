package campaign

import (
	"net/http"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	gsv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
)

type Handler struct {
	ucs UseCases
	gsv gsv.Server
}

func NewHandler(s gsv.Server, u UseCases) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
	}
}

func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/campaigns"

	public := router.Group(apiBase)
	{
		public.GET("", h.ListCampaigns)
	}
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	campaigns, err := h.ucs.ListCampaigns(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaigns)
}
