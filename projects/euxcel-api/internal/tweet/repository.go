package tweet

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	cass "github.com/devpablocristo/monorepo/pkg/databases/nosql/cassandra/gocql"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/repository/models"
	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// cassandra es la implementación del repositorio en Cassandra.
type cassandra struct {
	repository cass.Repository
}

// NewRepository retorna una implementación del repositorio para Cassandra.
func NewRepository(r cass.Repository) Repository {
	return &cassandra{
		repository: r,
	}
}

// SaveTweet inserta un tweet en la tabla global "tweets" de Cassandra.
func (r *cassandra) SaveTweet(ctx context.Context, tweet *domain.Tweet) (string, error) {
	if tweet == nil {
		return "", errors.New("tweet cannot be nil")
	}

	// Convertir del dominio al modelo de Cassandra.
	cassTweet, err := models.FromDomain(tweet)
	if err != nil {
		return "", fmt.Errorf("failed to convert tweet to Cassandra model: %w", err)
	}

	// Asignar un nuevo UUID y fecha de creación (si no está asignada).
	cassTweet.ID = uuid.New().String()
	if cassTweet.CreatedAt.IsZero() {
		cassTweet.CreatedAt = time.Now()
	}

	const query = "INSERT INTO tweets (id, user_id, content, created_at) VALUES (?, ?, ?, ?)"
	if err := r.repository.GetSession().Query(query,
		cassTweet.ID, cassTweet.UserID, cassTweet.Content, cassTweet.CreatedAt).
		WithContext(ctx).
		Exec(); err != nil {
		return "", fmt.Errorf("failed to save tweet: %w", err)
	}

	return cassTweet.ID, nil
}

// GetTweetsByUserIDs consulta los tweets de la vista desnormalizada "timeline_by_user"
// para una lista de userIDs. El parámetro offset se ignora, ya que Cassandra no lo soporta de forma nativa.
func (r *cassandra) GetTweetsByUserIDs(ctx context.Context, userIDs []string, limit int, offset int) ([]domain.Tweet, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("userIDs list is empty")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}

	// Consultar la tabla desnormalizada "timeline_by_user".
	// Se selecciona tweet_id (para asignarlo a ID), user_id, content y created_at.
	const query = "SELECT tweet_id, user_id, content, created_at FROM timeline_by_user WHERE user_id IN ? LIMIT ?"
	iter := r.repository.GetSession().Query(query, userIDs, limit).
		WithContext(ctx).
		Iter()

	var tweet models.Tweet
	var tweets []models.Tweet
	for iter.Scan(&tweet.ID, &tweet.UserID, &tweet.Content, &tweet.CreatedAt) {
		// Hacemos una copia local para evitar problemas de memoria.
		t := tweet
		tweets = append(tweets, t)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close iterator: %w", err)
	}

	// Convertir los modelos de Cassandra al objeto del dominio.
	domainTweets, err := models.ToDomainSlice(tweets)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tweets to domain model: %w", err)
	}

	// Ordenar los tweets de forma descendente por CreatedAt (más reciente primero).
	sort.Slice(domainTweets, func(i, j int) bool {
		return domainTweets[i].CreatedAt.After(domainTweets[j].CreatedAt)
	})

	return domainTweets, nil
}

// InsertTweetIntoTimeline inserta un tweet en la tabla desnormalizada "timeline_by_user",
// asociándolo al timeline del usuario (followee) pasado como parámetro.
func (r *cassandra) InsertTweetIntoTimeline(ctx context.Context, followeeID string, tweet *domain.Tweet) error {
	// Convertir del dominio al modelo de Cassandra.
	cassTweet, err := models.FromDomain(tweet)
	if err != nil {
		return fmt.Errorf("failed to convert tweet to Cassandra model: %w", err)
	}

	const query = `
		INSERT INTO timeline_by_user (user_id, created_at, tweet_id, content)
		VALUES (?, ?, ?, ?)
	`
	if err := r.repository.GetSession().Query(query,
		followeeID, cassTweet.CreatedAt, cassTweet.ID, cassTweet.Content).
		WithContext(ctx).
		Exec(); err != nil {
		return fmt.Errorf("failed to insert tweet into timeline for user %s: %w", followeeID, err)
	}
	return nil
}
