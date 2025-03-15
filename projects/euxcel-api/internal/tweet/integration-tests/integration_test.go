package integration_tests

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	rabbit "github.com/devpablocristo/monorepo/pkg/brokers/rabbitmq/amqp091/producer"
	redis "github.com/devpablocristo/monorepo/pkg/databases/cache/redis/v8"
	cass "github.com/devpablocristo/monorepo/pkg/databases/nosql/cassandra/gocql"
	gorm "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"
	pg "github.com/devpablocristo/monorepo/pkg/databases/sql/postgresql/pgxpool"

	// Usecases y dominio de tweets.
	tweet "github.com/devpablocristo/monorepo/projects/qh/internal/tweet"
	tweetDomain "github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"

	// Usecases de usuarios.
	user "github.com/devpablocristo/monorepo/projects/qh/internal/user"
	userDomain "github.com/devpablocristo/monorepo/projects/qh/internal/user/usecases/domain"

	// Usecases de personas.
	person "github.com/devpablocristo/monorepo/projects/qh/internal/person"
	personDomain "github.com/devpablocristo/monorepo/projects/qh/internal/person/usecases/domain"
)

func TestTweetWithUserAndPersonIntegration(t *testing.T) {
	// Cassandra: se espera que esté corriendo y configurado vía variables de entorno.
	cassandraRepo, err := cass.Bootstrap()
	assert.NoError(t, err, "Error bootstrapping Cassandra repository")
	tweetRepo := tweet.NewRepository(cassandraRepo)

	// Redis: se espera que esté corriendo y configurado.
	redisCache, err := redis.Bootstrap("", "", 0)
	assert.NoError(t, err, "Error bootstrapping Redis cache")
	tweetCache := tweet.NewCache(redisCache)

	// RabbitMQ: bootstrap del broker real usando variables de entorno.
	rabbitBroker, err := rabbit.Bootstrap()
	assert.NoError(t, err, "Error bootstrapping RabbitMQ broker")
	// Se asume que en el paquete tweet existe NewBroker que recibe el broker real y la routing key.
	tweetBroker := tweet.NewBroker(rabbitBroker, os.Getenv("RABBITMQ_ROUTING_KEY"))

	// Bootstrap del repositorio GORM para usuarios.
	userDB, err := gorm.Bootstrap("", "", "", "", "", 0)
	assert.NoError(t, err, "Error bootstrapping GORM repository for users")
	userRepo := user.NewRepository(userDB)
	userUseCases := user.NewUseCases(userRepo)

	personPool, _ := pg.Bootstrap("", "", "", "", "", "")
	personRepo := person.NewPostgresRepository(personPool)
	personUseCases := person.NewUseCases(personRepo)

	// --- Crear una persona ---
	newPerson := &personDomain.Person{
		FirstName:  "John",
		LastName:   "Doe",
		Age:        30,
		Gender:     "male",
		NationalID: time.Now().Unix(), // Genera un número único
		Phone:      "555-1234",
		Interests:  []string{"music", "sports"},
		Hobbies:    []string{"guitar", "running"},
	}
	personID, err := personUseCases.CreatePerson(context.Background(), newPerson)
	assert.NoError(t, err, "Error creating person")
	t.Logf("Person created with ID: %s", personID)

	// --- Crear un usuario vinculado a la persona ---
	newUser := &userDomain.User{
		PersonID:       personID,
		Credentials:    userDomain.Credentials{Email: "john.doe@example.com", Password: "secret"},
		UserType:       userDomain.UserTypePerson,
		EmailValidated: true,
	}
	userID, err := userUseCases.CreateUser(context.Background(), newUser)
	assert.NoError(t, err, "Error creating user")
	t.Logf("User created with ID: %s", userID)

	// --- Crear un tweet usando el ID del usuario ---
	tweetToCreate, err := tweetDomain.NewTweet(userID, "Hello Integration from user "+userID)
	assert.NoError(t, err, "Error creating domain tweet")

	// Crear la instancia de usecases para tweets.
	tweetUseCases := tweet.NewUseCases(tweetRepo, userUseCases, tweetCache, tweetBroker)
	createdTweetID, err := tweetUseCases.CreateTweet(context.Background(), tweetToCreate)
	assert.NoError(t, err, "CreateTweet returned an error")
	assert.NotEmpty(t, createdTweetID, "Expected non-empty tweet ID")
	t.Logf("Tweet created with ID: %s", createdTweetID)

	// --- Limpieza: eliminar el usuario y la persona ---
	err = userUseCases.DeleteUser(context.Background(), userID, true)
	assert.NoError(t, err, "Error deleting user")
	t.Log("User deleted successfully")

	err = personUseCases.DeletePerson(context.Background(), personID, true)
	assert.NoError(t, err, "Error deleting person")
	t.Log("Person deleted successfully")
}

func TestGetTimelineIntegration(t *testing.T) {

	cassRepo, err := cass.Bootstrap()
	assert.NoError(t, err, "Error bootstrapping Cassandra repository")
	tweetRepo := tweet.NewRepository(cassRepo)

	redisConn, err := redis.Bootstrap("", "", 0)
	assert.NoError(t, err, "Error bootstrapping Redis cache")
	tweetCache := tweet.NewCache(redisConn)

	rabbitBroker, err := rabbit.Bootstrap()
	assert.NoError(t, err, "Error bootstrapping RabbitMQ broker")
	tweetBroker := tweet.NewBroker(rabbitBroker, "")

	userDB, err := gorm.Bootstrap("", "", "", "", "", 0)
	assert.NoError(t, err, "Error bootstrapping GORM repository for users")
	userRepo := user.NewRepository(userDB)
	userUseCases := user.NewUseCases(userRepo)

	personPool, _ := pg.Bootstrap("", "", "", "", "", "")
	personRepo := person.NewPostgresRepository(personPool)
	personUseCases := person.NewUseCases(personRepo)

	newPerson := &personDomain.Person{
		FirstName:  "John",
		LastName:   "Doe",
		Age:        30,
		Gender:     "male",
		NationalID: time.Now().Unix(),
		Phone:      "555-1234",
		Interests:  []string{"music", "sports"},
		Hobbies:    []string{"guitar", "running"},
	}
	personID, err := personUseCases.CreatePerson(context.Background(), newPerson)
	assert.NoError(t, err, "Error creating person")
	t.Logf("Person created with ID: %s", personID)

	newUser := &userDomain.User{
		PersonID:       personID,
		Credentials:    userDomain.Credentials{Email: "john.doe@example.com", Password: "secret"},
		UserType:       userDomain.UserTypePerson,
		EmailValidated: true,
	}
	userID, err := userUseCases.CreateUser(context.Background(), newUser)
	assert.NoError(t, err, "Error creating user")
	t.Logf("User created with ID: %s", userID)

	var followerUserIDs []string
	for i := 0; i < 5; i++ {
		// Crear persona para cada seguidor.
		followerPerson := &personDomain.Person{
			FirstName:  "Follower" + strconv.Itoa(i+1),
			LastName:   "Test",
			Age:        25 + i,
			Gender:     "other",
			NationalID: time.Now().UnixNano(), // Valor único
			Phone:      "555-100" + strconv.Itoa(i),
			Interests:  []string{"reading", "coding"},
			Hobbies:    []string{"chess", "cycling"},
		}
		followerPersonID, err := personUseCases.CreatePerson(context.Background(), followerPerson)
		assert.NoError(t, err, "Error creating follower person")
		t.Logf("Follower person %d created with ID: %s", i+1, followerPersonID)

		// Crear usuario seguidor con email único.
		uniqueFollowerEmail := "follower" + strconv.Itoa(i+1) + "+" + time.Now().Format("150405") + "@example.com"
		followerUser := &userDomain.User{
			PersonID:       followerPersonID,
			Credentials:    userDomain.Credentials{Email: uniqueFollowerEmail, Password: "secret"},
			UserType:       userDomain.UserTypePerson,
			EmailValidated: true,
		}
		followerUserID, err := userUseCases.CreateUser(context.Background(), followerUser)
		assert.NoError(t, err, "Error creating follower user")
		t.Logf("Follower user %d created with ID: %s", i+1, followerUserID)
		followerUserIDs = append(followerUserIDs, followerUserID)

		// Crear la relación de seguimiento: cada seguidor sigue al usuario principal.
		followRel, err := userUseCases.FollowUser(context.Background(), followerUserID, userID)
		assert.NoError(t, err, "Error creating follow relationship for follower %d", i+1)
		t.Logf("Follower user %d follows main user, relation: %s", i+1, followRel)
	}

	tweetToCreate, err := tweetDomain.NewTweet(userID, "Hello Integration from user "+userID)
	assert.NoError(t, err, "Error creating domain tweet")

	tweetUseCases := tweet.NewUseCases(tweetRepo, userUseCases, tweetCache, tweetBroker)
	createdTweetID, err := tweetUseCases.CreateTweet(context.Background(), tweetToCreate)
	assert.NoError(t, err, "CreateTweet returned an error")
	assert.NotEmpty(t, createdTweetID, "Expected non-empty tweet ID")
	t.Logf("Tweet created with ID: %s", createdTweetID)

	timeline, err := tweetUseCases.GetTimeline(context.Background(), userID)
	assert.NoError(t, err, "GetTimeline returned an error")
	assert.NotEmpty(t, timeline, "Expected non-empty timeline")
	t.Logf("Timeline first tweet: %v", timeline[0])

	err = userUseCases.DeleteUser(context.Background(), userID, true)
	assert.NoError(t, err, "Error deleting user")
	t.Log("User deleted successfully")

	err = personUseCases.DeletePerson(context.Background(), personID, true)
	assert.NoError(t, err, "Error deleting person")
	t.Log("Person deleted successfully")
}
