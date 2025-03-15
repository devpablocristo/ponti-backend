package tweet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	redis0 "github.com/go-redis/redis/v8"

	redis "github.com/devpablocristo/monorepo/pkg/databases/cache/redis/v8"

	models "github.com/devpablocristo/monorepo/projects/qh/internal/tweet/cache/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// cache es la implementación de Cache utilizando Redis.
type cache struct {
	client redis.Cache
}

// NewRedisCache crea una nueva instancia de Cache basada en Redis.
func NewCache(c redis.Cache) Cache {
	return &cache{
		client: c,
	}
}

// InvalidateUserTimeline elimina la entrada de caché de la línea de tiempo de un usuario.
func (r *cache) InvalidateUserTimeline(ctx context.Context, userID string) error {
	key := fmt.Sprintf("timeline:%s", userID)
	if err := r.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete timeline cache for user %s: %w", userID, err)
	}
	return nil
}

// GetTimeline obtiene la línea de tiempo de un usuario desde Redis.
// Se espera que los datos estén almacenados en formato JSON.
func (r *cache) GetTimeline(ctx context.Context, userID string) ([]domain.Tweet, error) {
	key := fmt.Sprintf("timeline:%s", userID)

	// Obtener los datos de Redis.
	data, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis0.Nil) {
			return nil, fmt.Errorf("timeline not found for user %s: %w", userID, err)
		}
		return nil, fmt.Errorf("failed to retrieve timeline cache for user %s: %w", userID, err)
	}

	// Deserializar el JSON en un slice de modelos de cache.
	var timelineCache []models.Tweet
	if err := json.Unmarshal([]byte(data), &timelineCache); err != nil {
		return nil, fmt.Errorf("failed to parse timeline data for user %s: %w", userID, err)
	}

	// Convertir a objetos del dominio.
	domainTweets, err := models.ToDomainSlice(timelineCache)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cache models to domain models for user %s: %w", userID, err)
	}

	// Ordenar los tweets por CreatedAt, de modo que el más reciente aparezca primero.
	sort.Slice(domainTweets, func(i, j int) bool {
		return domainTweets[i].CreatedAt.After(domainTweets[j].CreatedAt)
	})

	return domainTweets, nil
}

// SetTimeline almacena la línea de tiempo de un usuario en Redis.
// Se serializa el slice de tweets a JSON y se guarda con una clave basada en el userID.
// Puedes agregar un tiempo de expiración si lo deseas.
func (r *cache) SetTimeline(ctx context.Context, userID string, tweets []domain.Tweet) error {
	key := fmt.Sprintf("timeline:%s", userID)

	// Convertir los tweets de dominio a modelos de cache (si es que manejas una conversión)
	timelineCache, err := models.FromDomainSlice(tweets)
	if err != nil {
		return fmt.Errorf("failed to convert domain models to cache models for user %s: %w", userID, err)
	}
	data, err := json.Marshal(timelineCache)
	if err != nil {
		return fmt.Errorf("failed to serialize timeline data for user %s: %w", userID, err)
	}

	// Guardar en Redis. El tercer parámetro es el tiempo de expiración;
	// en este ejemplo se deja sin expiración (0), pero podrías definir uno, por ejemplo, 1 hora.
	if err := r.client.Set(ctx, key, string(data), 0); err != nil {
		return fmt.Errorf("failed to set timeline cache for user %s: %w", userID, err)
	}

	return nil
}

// PushTweetToTimeline inserta el tweet en el timeline del usuario especificado.
// Se serializa el tweet a JSON y se utiliza la lista de Redis para almacenar el tweet.
// Además, se trunca la lista para mantener un tamaño máximo (por ejemplo, 100 tweets).
func (r *cache) PushTweetToTimeline(ctx context.Context, userID string, tweet *domain.Tweet) error {
	// Serializar el tweet a JSON.
	data, err := json.Marshal(tweet)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet: %w", err)
	}

	// Definir la llave del timeline del usuario.
	key := fmt.Sprintf("timeline:%s", userID)

	// Insertar el tweet al inicio de la lista.
	if err := r.client.LPush(ctx, key, data); err != nil {
		return fmt.Errorf("failed to push tweet to timeline: %w", err)
	}

	// Truncar la lista para mantener solo los últimos 100 tweets.
	if err := r.client.LTrim(ctx, key, 0, 99); err != nil {
		return fmt.Errorf("failed to trim timeline: %w", err)
	}

	return nil
}

// Close cierra la conexión con el cliente Redis.
func (r *cache) Close() {
	r.client.Close()
}
