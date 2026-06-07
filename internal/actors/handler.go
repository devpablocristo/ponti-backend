// Package actors expone los endpoints HTTP del registro de identidad (/actors/*).
package actors

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/actors/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	Resolve(context.Context, domain.ResolveInput) (domain.ResolveResult, error)
	GetByTaxID(context.Context, string) (*domain.Actor, error)
	Search(context.Context, string, int) (domain.SearchResult, error)
	List(context.Context, string, int, int) ([]domain.Actor, int64, error)
	Get(context.Context, int64) (*domain.Actor, error)
	Update(context.Context, *domain.Actor) error
	Archive(context.Context, int64) error
	Restore(context.Context, int64) error
	Delete(context.Context, int64) error
}

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{ucs: u, gsv: s, acf: c, mws: m}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/actors"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", h.ResolveActor)
		public.GET("", h.ListActors)
		public.GET("/search", h.SearchActors)
		public.GET("/by-tax-id", h.GetByTaxID)
		public.GET("/similar", h.SimilarActors)
		public.GET("/:actor_id", h.GetActor)
		public.PUT("/:actor_id", h.UpdateActor)
		public.DELETE("/:actor_id", h.DeleteActor)
		public.POST("/:actor_id/archive", h.ArchiveActor)
		public.POST("/:actor_id/restore", h.RestoreActor)
	}
}

// ResolveActor (POST /actors): resolve-or-create. 200 si reusó, 201 si creó.
func (h *Handler) ResolveActor(c *gin.Context) {
	var req dto.ResolveActorRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	res, err := h.ucs.Resolve(c.Request.Context(), req.ToDomain())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	payload := dto.ResolveFromDomain(res)
	if res.Reused {
		sharedhandlers.RespondOK(c, payload)
		return
	}
	ginmw.WriteJSON(c, http.StatusCreated, payload)
}

// SearchActors (GET /actors/search?q=&limit=): exactos + similares.
func (h *Handler) SearchActors(c *gin.Context) {
	res, err := h.ucs.Search(c.Request.Context(), c.Query("q"), parseLimit(c.Query("limit")))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.SearchFromDomain(res))
}

// GetByTaxID (GET /actors/by-tax-id?tax_id=): 200 actor | 404 | 422.
func (h *Handler) GetByTaxID(c *gin.Context) {
	taxID := strings.TrimSpace(c.Query("tax_id"))
	if taxID == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("tax_id required"))
		return
	}
	a, err := h.ucs.GetByTaxID(c.Request.Context(), taxID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.ActorFromDomain(a))
}

// SimilarActors (GET /actors/similar?name=&limit=): candidatos advisory (siempre 200).
func (h *Handler) SimilarActors(c *gin.Context) {
	q := c.Query("name")
	if q == "" {
		q = c.Query("q")
	}
	res, err := h.ucs.Search(c.Request.Context(), q, parseLimit(c.Query("limit")))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.CandidatesFromDomain(res))
}

func parseLimit(s string) int {
	if s == "" {
		return 20
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 20
	}
	if n > 50 {
		return 50
	}
	return n
}
