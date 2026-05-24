// Smoke test integración: ejercita el cliente Companion real contra el servicio
// levantado por `axis/docker-compose.yml` en localhost:18085.
//
// Run desde la raíz de ponti-backend:
//
//	go run ./scripts/smoke-companion
//
// Variables opcionales:
//
//	COMPANION_BASE_URL                (default http://localhost:18085)
//	COMPANION_INTERNAL_JWT_SECRET     (default axis-dev-internal-jwt-secret-change-me)
//	COMPANION_INTERNAL_JWT_ISSUER     (default axis-bff, matchea docker-compose)
//	COMPANION_INTERNAL_JWT_AUDIENCE   (default companion)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/devpablocristo/ponti-backend/internal/axis"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	client, err := axis.NewCompanionClient(axis.Config{
		BaseURL:     envOr("COMPANION_BASE_URL", "http://localhost:18085"),
		JWTSecret:   envOr("COMPANION_INTERNAL_JWT_SECRET", "axis-dev-internal-jwt-secret-change-me"),
		JWTIssuer:   envOr("COMPANION_INTERNAL_JWT_ISSUER", "axis-bff"),
		JWTAudience: envOr("COMPANION_INTERNAL_JWT_AUDIENCE", "companion"),
		Timeout:     10 * time.Second,
		MaxRetries:  0,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewCompanionClient: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	call := axis.CallContext{
		OrgID:  "local-dev-org",
		Actor:  "ponti-smoke@ponti.local",
		Scopes: []string{"companion:tasks:read", "companion:tasks:write"},
	}

	fmt.Println("=== POST /v1/chat ===")
	resp, err := client.Chat(ctx, call, axis.ChatRequest{
		Message:        "smoke test desde ponti-backend",
		Channel:        "api",
		ProductSurface: "ponti",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Chat error: %v\n", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))

	fmt.Println("\n=== GET /v1/chat/conversations?limit=5 ===")
	list, err := client.ListConversations(ctx, call, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ListConversations error: %v\n", err)
		os.Exit(1)
	}
	out, _ = json.MarshalIndent(list, "", "  ")
	fmt.Println(string(out))

	if resp.ChatID != "" {
		fmt.Printf("\n=== GET /v1/chat/conversations/%s ===\n", resp.ChatID)
		detail, err := client.GetConversation(ctx, call, resp.ChatID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "GetConversation error: %v\n", err)
			os.Exit(1)
		}
		out, _ = json.MarshalIndent(detail, "", "  ")
		fmt.Println(string(out))
	}

	fmt.Println("\n=== SMOKE TEST PASS ===")
}
