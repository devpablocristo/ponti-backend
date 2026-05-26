package axis

import (
	"errors"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// scopeClaim arma el claim `scope` (OAuth2 estándar) como string space-separated.
// Companion lo lee así en `wire/auth.go::claimScopes` (espera `scope` o `scp`,
// no `scopes`). Mantenemos `scopes` también por si algún consumer espera array.
func scopeClaim(scopes []string) string {
	return strings.Join(scopes, " ")
}

// jwtSigner firma JWTs internos HS256 que Companion (y Nexus, cuando se integre)
// validan contra el mismo secret compartido (`COMPANION_INTERNAL_JWT_SECRET`).
//
// Por qué JWT y no API key:
// - Companion sanitiza los headers `X-Org-ID`/`X-User-ID` con el principal
//   resuelto en el middleware de auth (anti-spoofing en
//   `platform/authn/go/identityhttp.WithPrincipal`). Con API key, el
//   principal queda fijo en el metadata de la key → no se puede impersonar
//   otro tenant.
// - Con JWT interno, los claims (`org_id`, `actor`, `scopes`) viajan firmados.
//   Companion los valida y respeta. Patrón idéntico al que usa `axis-bff` para
//   firmar tokens por usuario.
//
// El secret se sincroniza vía Secret Manager GCP en prod y env var en local.
type jwtSigner struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

// newJWTSigner construye el firmer. Falla si el secret está vacío para no
// emitir tokens inválidos silenciosamente.
func newJWTSigner(secret, issuer, audience string, ttl time.Duration) (*jwtSigner, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("axis: companion internal jwt secret required")
	}
	if strings.TrimSpace(issuer) == "" {
		issuer = "ponti-backend"
	}
	if strings.TrimSpace(audience) == "" {
		audience = "companion"
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &jwtSigner{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}, nil
}

// sign emite un JWT HS256 con `iss`, `aud`, `sub`, `exp`, `iat` estándar +
// claims `org_id`, `actor`, `scopes` que Companion lee del JWT interno.
//
// Companion exige `sub` (o `actor_id`) en `internal_jwt.go:79-82` para resolver
// el principal — sin esto rechaza con `missing internal jwt subject`. Seteamos
// `sub = actor` y también `actor_id` para compatibilidad.
//
// El token es short-lived (ttl default 5min) para minimizar el blast radius si
// se filtra. No hay refresh; cada request firma un token nuevo.
func (s *jwtSigner) sign(call CallContext) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      s.issuer,
		"aud":      s.audience,
		"sub":      call.Actor,
		"iat":      now.Unix(),
		"exp":      now.Add(s.ttl).Unix(),
		"org_id":   call.OrgID,
		"actor":    call.Actor,
		"actor_id": call.Actor,
		"scope":    scopeClaim(call.Scopes), // OAuth2 standard, leído por Companion
		"scopes":   call.Scopes,              // legacy, por si algo lo espera como array
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}
	return signed, nil
}
