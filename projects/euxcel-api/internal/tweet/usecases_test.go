package tweet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mock_tweet "github.com/devpablocristo/monorepo/projects/qh/internal/tweet/mocks"
	mock_user "github.com/devpablocristo/monorepo/projects/qh/internal/user/mocks"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
	usrdom "github.com/devpablocristo/monorepo/projects/qh/internal/user/usecases/domain"
)

func TestCreateTweet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Estructura para definir las dependencias mockeadas.
	type fields struct {
		cassRepository *mock_tweet.MockRepository
		userUC         *mock_user.MockUseCases
		cache          *mock_tweet.MockCache
		producer       *mock_tweet.MockBroker
	}
	type args struct {
		ctx   context.Context
		tweet *domain.Tweet
	}
	tests := []struct {
		name        string
		setup       func(f *fields)
		args        args
		wantErr     bool
		wantTweetID string
	}{
		{
			name: "Error: User not found",
			setup: func(f *fields) {
				// Se espera que se llame a GetUser y retorne (nil, nil)
				f.userUC.EXPECT().
					GetUser(gomock.Any(), "user1").
					Return(nil, nil)
			},
			args: args{
				ctx:   context.Background(),
				tweet: &domain.Tweet{UserID: "user1", Content: "Hello World"},
			},
			wantErr: true,
		},
		{
			name: "Error: Failed to save tweet to Cassandra",
			setup: func(f *fields) {
				// Simular usuario encontrado.
				f.userUC.EXPECT().
					GetUser(gomock.Any(), "user1").
					Return(&usrdom.User{ID: "user1"}, nil)
				// Simular error al guardar el tweet.
				f.cassRepository.EXPECT().
					SaveTweet(gomock.Any(), gomock.Any()).
					Return("", errors.New("error saving tweet"))
			},
			args: args{
				ctx:   context.Background(),
				tweet: &domain.Tweet{UserID: "user1", Content: "Hello World"},
			},
			wantErr: true,
		},
		{
			name: "Success: Tweet creation successful",
			setup: func(f *fields) {
				// 1. El usuario existe.
				f.userUC.EXPECT().
					GetUser(gomock.Any(), "user1").
					Return(&usrdom.User{ID: "user1"}, nil)
				// 2. Guardar tweet en Cassandra devuelve un ID.
				f.cassRepository.EXPECT().
					SaveTweet(gomock.Any(), gomock.Any()).
					Return("tweet123", nil)
				// 3. Obtener followers del usuario.
				f.userUC.EXPECT().
					GetFollowerUsers(gomock.Any(), "user1").
					Return([]string{"follower1", "follower2"}, nil)
				// 4. Invalida la caché de cada follower.
				f.cache.EXPECT().
					InvalidateUserTimeline(gomock.Any(), "follower1").
					Return(nil)
				f.cache.EXPECT().
					InvalidateUserTimeline(gomock.Any(), "follower2").
					Return(nil)
				// 5. Se inserta el tweet en el timeline de cada follower.
				f.cassRepository.EXPECT().
					InsertTweetIntoTimeline(gomock.Any(), "follower1", gomock.Any()).
					Return(nil)
				f.cassRepository.EXPECT().
					InsertTweetIntoTimeline(gomock.Any(), "follower2", gomock.Any()).
					Return(nil)
				// 6. Se publica el evento de tweet creado.
				f.producer.EXPECT().
					PublishTweetCreated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			args: args{
				ctx:   context.Background(),
				tweet: &domain.Tweet{UserID: "user1", Content: "Hello World"},
			},
			wantErr:     false,
			wantTweetID: "tweet123",
		},
		{
			name: "Success: Tweet creation with error in one InsertTweetIntoTimeline",
			setup: func(f *fields) {
				// Simular que el usuario existe.
				f.userUC.EXPECT().
					GetUser(gomock.Any(), "user1").
					Return(&usrdom.User{ID: "user1"}, nil)
				// Guardar tweet en Cassandra devuelve un ID.
				f.cassRepository.EXPECT().
					SaveTweet(gomock.Any(), gomock.Any()).
					Return("tweet123", nil)
				// Obtener followers del usuario.
				f.userUC.EXPECT().
					GetFollowerUsers(gomock.Any(), "user1").
					Return([]string{"follower1", "follower2"}, nil)
				// Invalida la caché de cada follower.
				f.cache.EXPECT().
					InvalidateUserTimeline(gomock.Any(), "follower1").
					Return(nil)
				f.cache.EXPECT().
					InvalidateUserTimeline(gomock.Any(), "follower2").
					Return(nil)
				// Insertar el tweet en el timeline:
				// Para "follower1" retorna nil, para "follower2" simula un error.
				f.cassRepository.EXPECT().
					InsertTweetIntoTimeline(gomock.Any(), "follower1", gomock.Any()).
					Return(nil)
				f.cassRepository.EXPECT().
					InsertTweetIntoTimeline(gomock.Any(), "follower2", gomock.Any()).
					Return(errors.New("error inserting timeline"))
				// Publicar el evento de tweet creado.
				f.producer.EXPECT().
					PublishTweetCreated(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			args: args{
				ctx:   context.Background(),
				tweet: &domain.Tweet{UserID: "user1", Content: "Hello World"},
			},
			// Aunque una inserción falle, el error se registra y se continúa el flujo.
			wantErr:     false,
			wantTweetID: "tweet123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Inicializar mocks.
			f := fields{
				cassRepository: mock_tweet.NewMockRepository(ctrl),
				userUC:         mock_user.NewMockUseCases(ctrl),
				cache:          mock_tweet.NewMockCache(ctrl),
				producer:       mock_tweet.NewMockBroker(ctrl),
			}
			// Configurar expectativas según el test case.
			tc.setup(&f)

			// Crear la instancia de usecases con las dependencias mockeadas.
			uc := NewUseCases(f.cassRepository, f.userUC, f.cache, f.producer)
			gotTweetID, err := uc.CreateTweet(tc.args.ctx, tc.args.tweet)

			// Usar assert de Testify para validar el resultado.
			if tc.wantErr {
				assert.Error(t, err, "expected an error but got nil")
			} else {
				assert.NoError(t, err, "expected no error but got one")
				assert.Equal(t, tc.wantTweetID, gotTweetID, "tweet ID mismatch")
			}
		})
	}
}

func TestGetTimeline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cassRepository *mock_tweet.MockRepository
		userUC         *mock_user.MockUseCases // No se usa, pero se debe inyectar
		cache          *mock_tweet.MockCache
		producer       *mock_tweet.MockBroker // Tampoco se usa aquí
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name       string
		setup      func(f *fields)
		args       args
		wantErr    bool
		wantTweets []domain.Tweet
	}{
		{
			name: "Cache hit: timeline obtained from cache",
			setup: func(f *fields) {
				// Simular que la caché retorna un timeline no vacío.
				tweets := []domain.Tweet{
					{ID: "tweet1", UserID: "user1", Content: "Message", CreatedAt: time.Now()},
				}
				f.cache.EXPECT().
					GetTimeline(gomock.Any(), "user1").
					Return(tweets, nil)
			},
			args: args{
				ctx:    context.Background(),
				userID: "user1",
			},
			wantErr:    false,
			wantTweets: []domain.Tweet{{ID: "tweet1", UserID: "user1", Content: "Message"}},
		},
		{
			name: "Cache miss: timeline retrieved from Cassandra and cache updated",
			setup: func(f *fields) {
				// Simular que la caché no tiene datos.
				f.cache.EXPECT().
					GetTimeline(gomock.Any(), "user1").
					Return(nil, nil)
				// Se consulta a Cassandra y se devuelve un timeline.
				tweets := []domain.Tweet{
					{ID: "tweet2", UserID: "user1", Content: "Another Message", CreatedAt: time.Now()},
				}
				f.cassRepository.EXPECT().
					GetTweetsByUserIDs(gomock.Any(), []string{"user1"}, 50, 0).
					Return(tweets, nil)
				// Actualización asíncrona de la caché.
				f.cache.EXPECT().
					SetTimeline(gomock.Any(), "user1", tweets).
					Return(nil)
			},
			args: args{
				ctx:    context.Background(),
				userID: "user1",
			},
			wantErr:    false,
			wantTweets: []domain.Tweet{{ID: "tweet2", UserID: "user1", Content: "Another Message"}},
		},
		{
			name: "Error retrieving timeline from Cassandra",
			setup: func(f *fields) {
				f.cache.EXPECT().
					GetTimeline(gomock.Any(), "user1").
					Return(nil, nil)
				f.cassRepository.EXPECT().
					GetTweetsByUserIDs(gomock.Any(), []string{"user1"}, 50, 0).
					Return(nil, errors.New("db error"))
			},
			args: args{
				ctx:    context.Background(),
				userID: "user1",
			},
			wantErr: true,
		},
		{
			name: "Cache error but timeline retrieved from Cassandra successfully",
			setup: func(f *fields) {
				f.cache.EXPECT().
					GetTimeline(gomock.Any(), "user1").
					Return(nil, errors.New("cache error"))
				tweets := []domain.Tweet{
					{ID: "tweet3", UserID: "user1", Content: "Alternate Message", CreatedAt: time.Now()},
				}
				f.cassRepository.EXPECT().
					GetTweetsByUserIDs(gomock.Any(), []string{"user1"}, 50, 0).
					Return(tweets, nil)
				f.cache.EXPECT().
					SetTimeline(gomock.Any(), "user1", tweets).
					Return(nil)
			},
			args: args{
				ctx:    context.Background(),
				userID: "user1",
			},
			wantErr:    false,
			wantTweets: []domain.Tweet{{ID: "tweet3", UserID: "user1", Content: "Alternate Message"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := fields{
				cassRepository: mock_tweet.NewMockRepository(ctrl),
				userUC:         mock_user.NewMockUseCases(ctrl),
				cache:          mock_tweet.NewMockCache(ctrl),
				producer:       mock_tweet.NewMockBroker(ctrl),
			}
			tc.setup(&f)
			uc := NewUseCases(f.cassRepository, f.userUC, f.cache, f.producer)
			gotTweets, err := uc.GetTimeline(tc.args.ctx, tc.args.userID)

			if tc.wantErr {
				assert.Error(t, err, "expected an error but got nil")
			} else {
				assert.NoError(t, err, "expected no error but got one")
				assert.Equal(t, len(tc.wantTweets), len(gotTweets), "number of tweets mismatch")
				// Opcional: comparar campos deterministas de los tweets.
			}
		})
	}
}
