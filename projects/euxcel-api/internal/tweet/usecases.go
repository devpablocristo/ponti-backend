package tweet

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
	"github.com/devpablocristo/monorepo/projects/qh/internal/user"
)

// usecases implementa la lógica de negocio para tweets.
type usecases struct {
	cassRepository Repository // Cassandra
	userUC         user.UseCases
	cache          Cache  // Redis
	producer       Broker // RabbitMQ
}

// NewUseCases crea una nueva instancia de usecases.
func NewUseCases(repo Repository, userUC user.UseCases, cache Cache, broker Broker) UseCases {
	return &usecases{
		cassRepository: repo,
		userUC:         userUC,
		cache:          cache,
		producer:       broker,
	}
}

func (uc *usecases) CreateTweet(ctx context.Context, tweet *domain.Tweet) (string, error) {
	// 1. Verificar que el autor existe.
	author, err := uc.userUC.GetUser(ctx, tweet.UserID)
	if err != nil {
		return "", fmt.Errorf("error trying to get user id %s: %w", tweet.UserID, err)
	}
	if author == nil {
		return "", fmt.Errorf("user id %s does not exist", tweet.UserID)
	}

	// 2. Crear el tweet en el dominio usando el constructor que asigna CreatedAt.
	newTweet, err := domain.NewTweet(tweet.UserID, tweet.Content)
	if err != nil {
		return "", err
	}

	// 3. Guardar el tweet en la tabla global "tweets".
	newTweetID, err := uc.cassRepository.SaveTweet(ctx, newTweet)
	if err != nil {
		return "", err
	}
	newTweet.ID = newTweetID

	// 4. Obtener la lista de seguidores (followers) del autor.
	followers, err := uc.userUC.GetFollowerUsers(ctx, newTweet.UserID)
	if err != nil {
		log.Printf("failed to get follower users for user %s: %v", newTweet.UserID, err)
		// Podemos decidir continuar sin fan‑out o retornar el error, según el caso de uso.
	}

	// 5. Invalida la caché del timeline de cada seguidor.
	for _, followerID := range followers {
		if err := uc.cache.InvalidateUserTimeline(ctx, followerID); err != nil {
			log.Printf("failed to invalidate timeline cache for follower %s: %v", followerID, err)
		}
	}

	// 6. Fan‑out: Insertar el tweet en la vista desnormalizada (timeline_by_user)
	//    de cada seguidor utilizando un worker pool.
	const numWorkers = 50
	followerChan := make(chan string, len(followers))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for followerID := range followerChan {
				// Insertar el tweet en el timeline del seguidor.
				if err := uc.cassRepository.InsertTweetIntoTimeline(ctx, followerID, newTweet); err != nil {
					log.Printf("error pushing tweet to timeline for follower %s: %v", followerID, err)
				}
			}
		}()
	}

outer:
	for _, followerID := range followers {
		select {
		case <-ctx.Done():
			log.Printf("context cancelled while sending follower IDs: %v", ctx.Err())
			break outer
		default:
			followerChan <- followerID
		}
	}
	close(followerChan)
	wg.Wait()

	// 7. Publicar un evento de tweet creado.
	if err := uc.producer.PublishTweetCreated(ctx, newTweet); err != nil {
		return "", fmt.Errorf("failed to publish tweet creation event: %v", err)
	}

	return newTweet.ID, nil
}

// GetTimeline consulta el timeline de un usuario. Primero intenta obtenerlo
// del caché; si no está, lo consulta en Cassandra (tabla desnormalizada, e.g. timeline_by_user)
// y actualiza el caché de forma asíncrona.
func (uc *usecases) GetTimeline(ctx context.Context, userID string) ([]domain.Tweet, error) {
	// 1. Intentar obtener el timeline desde la caché.
	cachedTimeline, err := uc.cache.GetTimeline(ctx, userID)
	if err == nil && cachedTimeline != nil && len(cachedTimeline) > 0 {
		return cachedTimeline, nil
	} else if err != nil {
		log.Printf("failed to get timeline from cache for user %s: %v", userID, err)
	}

	// 2. Consultar el timeline en Cassandra.
	// Se pasa el userID en un slice, ya que el repositorio espera []string.
	tweets, err := uc.cassRepository.GetTweetsByUserIDs(ctx, []string{userID}, 50, 0)
	if err != nil {
		return nil, fmt.Errorf("error retrieving timeline from Cassandra: %w", err)
	}

	// 3. Actualizar el caché de forma asíncrona.
	errCh := make(chan error, 1)
	go func() {
		updateCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := uc.cache.SetTimeline(updateCtx, userID, tweets); err != nil {
			errCh <- fmt.Errorf("failed to cache timeline for user %s: %w", userID, err)
			return
		}
		errCh <- nil
	}()

	select {
	case cacheErr := <-errCh:
		if cacheErr != nil {
			log.Printf("cache update error for user %s: %v", userID, cacheErr)
		}
	case <-time.After(1 * time.Second):
		log.Printf("cache update for user %s is taking longer than expected", userID)
	}

	return tweets, nil
}
