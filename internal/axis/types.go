// Package axis contiene el cliente HTTP para hablar con Companion (y, más adelante,
// con Nexus). El paquete encapsula la firma de JWTs internos, la traducción de
// errores HTTP a domainerr y los DTOs que matchean el OpenAPI de Companion en
// `axis/companion/openapi.yaml`.
package axis

import "time"

// ChatRequest matchea el DTO de Companion en
// `axis/companion/internal/tasks/handler/dto/dto.go::ChatRequest`.
// Para continuar una conversación durable, Ponti debe enviar `chat_id`; `task_id`
// queda reservado para callers internos que conocen el UUID de la task.
type ChatRequest struct {
	Message        string `json:"message"`
	TaskID         string `json:"task_id,omitempty"`
	ChatID         string `json:"chat_id,omitempty"`
	Channel        string `json:"channel,omitempty"`
	ProductSurface string `json:"product_surface,omitempty"`
}

// ChatResponse matchea el schema `ChatResponse` del OpenAPI de Companion.
type ChatResponse struct {
	ChatID               string         `json:"chat_id,omitempty"`
	TaskID               string         `json:"task_id,omitempty"`
	Reply                string         `json:"reply"`
	Blocks               []ChatBlock    `json:"blocks,omitempty"`
	ToolCalls            []ChatToolCall `json:"tool_calls,omitempty"`
	PendingConfirmations []any          `json:"pending_confirmations,omitempty"`
	Task                 Task           `json:"task"`
	Messages             []Message      `json:"messages"`
	RoutedAgent          string         `json:"routed_agent,omitempty"`
	RoutingSource        string         `json:"routing_source,omitempty"`
}

// ChatBlock es un bloque del response (texto, código, etc.). Companion los usa
// para componer respuestas estructuradas.
type ChatBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ChatToolCall describe una herramienta invocada por el agente durante el turno.
type ChatToolCall struct {
	Name   string         `json:"name"`
	Args   map[string]any `json:"args,omitempty"`
	Result map[string]any `json:"result,omitempty"`
}

// Task es la entidad operativa que Companion crea para cada chat.
// Mantiene estado durable de la conversación.
type Task struct {
	ID          string         `json:"id"`
	OrgID       string         `json:"org_id"`
	Title       string         `json:"title"`
	Goal        string         `json:"goal,omitempty"`
	Status      string         `json:"status"`
	Priority    string         `json:"priority,omitempty"`
	CreatedBy   string         `json:"created_by,omitempty"`
	AssignedTo  string         `json:"assigned_to,omitempty"`
	Channel     string         `json:"channel,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	ContextJSON map[string]any `json:"context_json,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Message es un mensaje individual dentro de una conversación.
type Message struct {
	ID         string         `json:"id"`
	AuthorType string         `json:"author_type"`
	AuthorID   string         `json:"author_id"`
	Body       string         `json:"body"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ConversationList es el response de `GET /v1/chat/conversations`.
type ConversationList struct {
	Items []ConversationSummary `json:"items"`
}

// ConversationSummary describe una conversación en el listado (sin mensajes).
type ConversationSummary struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
	MessageCount   int       `json:"message_count"`
	ProductSurface string    `json:"product_surface,omitempty"`
}

// ConversationMessage describe un mensaje persistido en agent_conversations.
// Es el contrato canónico de Companion/platform para el historial.
type ConversationMessage struct {
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	Blocks    []ChatBlock `json:"blocks,omitempty"`
}

// ConversationDetail es el response de `GET /v1/chat/conversations/{id}`.
// Incluye los mensajes ordenados cronológicamente.
type ConversationDetail struct {
	ID             string                `json:"id"`
	Title          string                `json:"title"`
	UpdatedAt      time.Time             `json:"updated_at"`
	CreatedAt      time.Time             `json:"created_at"`
	MessageCount   int                   `json:"message_count,omitempty"`
	ProductSurface string                `json:"product_surface,omitempty"`
	Messages       []ConversationMessage `json:"messages"`
}

// CallContext lleva la identidad real del usuario final que origina el turno.
// El cliente firma un JWT corto por request con estos valores como claims.
// Companion los respeta porque sanitiza headers pero confía en claims JWT
// internos firmados con el secret compartido.
type CallContext struct {
	OrgID  string   // tenant_id de Ponti (UUID en string)
	Actor  string   // email o id del local_user que dispara el turno
	Scopes []string // scopes companion:* requeridos para la operación
}
