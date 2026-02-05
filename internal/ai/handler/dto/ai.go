package dto

type AskContext struct {
	DateFrom string `json:"date_from,omitempty"`
	DateTo   string `json:"date_to,omitempty"`
}

type AskRequest struct {
	Question string      `json:"question"`
	Context  *AskContext `json:"context,omitempty"`
}

type AskResponse struct {
	RequestID string           `json:"request_id"`
	Intent    string           `json:"intent"`
	QueryID   *string          `json:"query_id"`
	Params    map[string]any   `json:"params"`
	Data      []map[string]any `json:"data"`
	Answer    string           `json:"answer"`
	Sources   []map[string]any `json:"sources"`
	Warnings  []string         `json:"warnings"`
}

type IngestDocument struct {
	Source   string         `json:"source"`
	Title    string         `json:"title"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type IngestRequest struct {
	Documents []IngestDocument `json:"documents"`
}

type IngestResponse struct {
	RequestID string `json:"request_id"`
	Ingested  int    `json:"ingested"`
}

type InsightItem struct {
	ID              string         `json:"id"`
	ProjectID       string         `json:"project_id"`
	EntityType      string         `json:"entity_type"`
	EntityID        string         `json:"entity_id"`
	Type            string         `json:"type"`
	Severity        int            `json:"severity"`
	Priority        int            `json:"priority"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary"`
	Evidence        map[string]any `json:"evidence"`
	Explanations    map[string]any `json:"explanations"`
	Action          map[string]any `json:"action"`
	ModelVersion    string         `json:"model_version"`
	FeaturesVersion string         `json:"features_version"`
	ComputedAt      string         `json:"computed_at"`
	ValidUntil      string         `json:"valid_until"`
	Status          string         `json:"status"`
}

type InsightsSummaryResponse struct {
	NewCountTotal        int           `json:"new_count_total"`
	NewCountHighSeverity int           `json:"new_count_high_severity"`
	TopInsights          []InsightItem `json:"top_insights"`
}

type InsightsListResponse struct {
	Insights []InsightItem `json:"insights"`
}

type ComputeInsightsResponse struct {
	RequestID       string `json:"request_id"`
	Computed        int    `json:"computed"`
	InsightsCreated int    `json:"insights_created"`
}

type ActionRequest struct {
	Action    string `json:"action"`
	NewStatus string `json:"new_status"`
}

type ActionResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

type JobRecomputeRequest struct {
	BatchSize *int `json:"batch_size,omitempty"`
}

type JobRecomputeBaselinesRequest struct {
	BatchSize *int `json:"batch_size,omitempty"`
}
