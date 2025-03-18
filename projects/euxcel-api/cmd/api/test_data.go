package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	person "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/person"
	personDomain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/person/usecases/domain"
	tweet "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/tweet"
	tweetDomain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/tweet/usecases/domain"
	user "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/user"
	userDomain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/user/usecases/domain"
)

// PersonUserSeed agrupa los datos de seed para una persona y su usuario asociado.
type PersonUserSeed struct {
	Person        personDomain.Person
	User          userDomain.User
	CreatedUserID string // Se almacena el ID del usuario creado
}

func seedTestData(ctx context.Context, pc person.UseCases, uc user.UseCases, tc tweet.UseCases) error {
	// Definición de los seeds (los registros de prueba).
	seeds := []PersonUserSeed{
		{
			Person: personDomain.Person{
				FirstName:  "Marge",
				LastName:   "Simpson",
				Age:        36,
				Gender:     "female",
				NationalID: 1234567891,
				Phone:      "+11234567891",
				Interests:  []string{"family", "cooking", "gardening"},
				Hobbies:    []string{"painting", "volunteering"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "marge.simpson@example.com",
					Password: "marge123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Bart",
				LastName:   "Simpson",
				Age:        10,
				Gender:     "male",
				NationalID: 1234567892,
				Phone:      "+11234567892",
				Interests:  []string{"skateboarding", "pranks", "troublemaking"},
				Hobbies:    []string{"skateboarding", "graffiti"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "bart.simpson@example.com",
					Password: "bart123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: false,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Lisa",
				LastName:   "Simpson",
				Age:        8,
				Gender:     "female",
				NationalID: 1234567893,
				Phone:      "+11234567893",
				Interests:  []string{"saxophone", "reading", "environment"},
				Hobbies:    []string{"playing saxophone", "studying", "volunteering"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "lisa.simpson@example.com",
					Password: "lisa123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Maggie",
				LastName:   "Simpson",
				Age:        1,
				Gender:     "female",
				NationalID: 1234567894,
				Phone:      "+11234567894",
				Interests:  []string{"crawling", "playing", "toys"},
				Hobbies:    []string{"crawling", "observing"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "maggie.simpson@example.com",
					Password: "maggie123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: false,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Ned",
				LastName:   "Flanders",
				Age:        60,
				Gender:     "male",
				NationalID: 1234567895,
				Phone:      "+11234567895",
				Interests:  []string{"baking", "praying", "gardening"},
				Hobbies:    []string{"gardening", "church activities"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "ned.flanders@example.com",
					Password: "ned123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Moe",
				LastName:   "Szyslak",
				Age:        45,
				Gender:     "male",
				NationalID: 1234567896,
				Phone:      "+11234567896",
				Interests:  []string{"drinking", "bar management"},
				Hobbies:    []string{"mixing cocktails", "arguing with customers"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "moe.szyslak@example.com",
					Password: "moe123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: false,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Barney",
				LastName:   "Gomez",
				Age:        40,
				Gender:     "male",
				NationalID: 1234567897,
				Phone:      "+11234567897",
				Interests:  []string{"beer", "napping", "singing"},
				Hobbies:    []string{"drinking", "sleeping"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "barney.gumble@example.com",
					Password: "barney123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: false,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Krusty",
				LastName:   "Krustofsky",
				Age:        55,
				Gender:     "male",
				NationalID: 1234567898,
				Phone:      "+11234567898",
				Interests:  []string{"comedy", "television", "entertainment"},
				Hobbies:    []string{"juggling", "performing"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "krusty@example.com",
					Password: "krusty123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Jefe",
				LastName:   "Gorgory",
				Age:        50,
				Gender:     "male",
				NationalID: 1234567899,
				Phone:      "+11234567899",
				Interests:  []string{"law enforcement", "donuts", "food"},
				Hobbies:    []string{"eating", "napping", "investigating"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "jefe.gorgory@example.com",
					Password: "jefe123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: false,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Montgomery",
				LastName:   "Burns",
				Age:        95,
				Gender:     "male",
				NationalID: 1234567900,
				Phone:      "+11234567900",
				Interests:  []string{"power", "wealth", "control"},
				Hobbies:    []string{"scheming", "investing", "hoarding money"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "montgomery.burns@example.com",
					Password: "burns123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Waylon",
				LastName:   "Smithers",
				Age:        40,
				Gender:     "male",
				NationalID: 1234567901,
				Phone:      "+11234567901",
				Interests:  []string{"loyalty", "management", "assistantship"},
				Hobbies:    []string{"serving", "administering", "planning"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "waylon.smithers@example.com",
					Password: "smithers123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
		{
			Person: personDomain.Person{
				FirstName:  "Seymour",
				LastName:   "Skinner",
				Age:        50,
				Gender:     "male",
				NationalID: 1234567902,
				Phone:      "+11234567902",
				Interests:  []string{"education", "discipline", "order"},
				Hobbies:    []string{"managing", "teaching", "organizing"},
			},
			User: userDomain.User{
				Credentials: userDomain.Credentials{
					Email:    "seymour.skinner@example.com",
					Password: "skinner123",
				},
				UserType:       userDomain.UserTypePerson,
				Roles:          []userDomain.Role{{Name: "user", Permissions: []userDomain.Permission{}}},
				LoggedAt:       time.Now(),
				EmailValidated: true,
			},
		},
	}

	// Crear un slice para almacenar los usuarios creados (con su ID).
	createdUsers := make([]PersonUserSeed, 0, len(seeds))

	// Por cada seed se crea la persona y luego se usa el ID retornado para crear el usuario.
	for _, seed := range seeds {
		// Crear la persona y obtener su ID.
		personID, err := pc.CreatePerson(ctx, &seed.Person)
		if err != nil {
			return fmt.Errorf("error creating person '%s %s': %v", seed.Person.FirstName, seed.Person.LastName, err)
		}
		// Asignar el ID retornado al usuario.
		seed.User.PersonID = personID

		createdUserID, err := uc.CreateUser(ctx, &seed.User)
		if err != nil {
			log.Printf("Error creating user with email '%s': %v", seed.User.Credentials.Email, err)
		} else {
			seed.CreatedUserID = createdUserID
			createdUsers = append(createdUsers, seed)
		}
	}

	// Buscar el ID de Marge (se asume que su FirstName es "Marge")
	var margeID string
	for _, u := range createdUsers {
		if u.Person.FirstName == "Marge" {
			margeID = u.CreatedUserID
			break
		}
	}
	if margeID == "" {
		return fmt.Errorf("margeID not found")
	}

	// Hacer que todos los usuarios (excepto Marge) sigan a Marge.
	for _, u := range createdUsers {
		if u.CreatedUserID != margeID {
			_, err := uc.FollowUser(ctx, u.CreatedUserID, margeID)
			if err != nil {
				log.Printf("Error: user %s could not follow Marge: %v", u.CreatedUserID, err)
			}
		}
	}

	// Hacer que Marge siga a todos los usuarios (excepto ella misma).
	for _, u := range createdUsers {
		if u.CreatedUserID != margeID {
			_, err := uc.FollowUser(ctx, margeID, u.CreatedUserID)
			if err != nil {
				log.Printf("Error: Marge could not follow user %s: %v", u.CreatedUserID, err)
			}
		}
	}

	// ========================
	// Creación de tweets con frases típicas asignadas a cada personaje
	// ========================
	// Definimos un map donde la clave es el nombre del personaje y el valor son sus frases características.
	characterPhrases := map[string][]string{
		"Marge":      {"¡Oh, cielos!", "¡Homero, por favor, sé responsable!", "¡La familia es lo primero!"},
		"Bart":       {"¡Ay caramba!", "¡Yo no fui!", "No te puedo prometer que lo trataré, pero trataré de tratarlo."},
		"Lisa":       {"Cualquier hombre que envidie a nuestra familia, es un hombre que necesita ayuda.", "¡La educación lo es todo!", "¡No me llames sabelotodo!"},
		"Maggie":     {"(murmullos de bebé)", "¡Gugu, ga ga!", "¡(silencio adorable)"},
		"Ned":        {"¡Okily Dokily!", "¡Dios me bendiga!", "¡Siempre con una sonrisa!"},
		"Moe":        {"¿Qué te pasa, amigo?", "Vas por la vida tratando de ser bueno con la gente, tratas de resistirte ante la tentación de golpearlos en la cara, ¿y todo para qué?", "Me llamo Moe, o como a las chicas les gusta decirme, “Ey, tú, el que está detrás de los arbustos”."},
		"Barney":     {"¡Eh, ¿qué pasó, amigo?!", "¡Mmm... cerveza!", "¡Salud, por la birra!"},
		"Krusty":     {"¡Ja, ja, ja! ¡Qué espectáculo!", "¡Bienvenidos a mi show!", "¡El circo ha llegado a Springfield!"},
		"Jefe":       {"¿Por qué la gente no puede hacer valer la ley con sus propias manos? Es decir, nosotros no podemos andar vigilando toda la ciudad.", "¡Detengan a esos maleantes!", "¡La ley es la ley!"},
		"Montgomery": {"¡Excelente!", "¿La causa de muerte de mis padres? Se interpusieron en mi camino.", "Familia, religión y amistad. Estos son tres demonios con los que hay que acabar si desean ser exitosos en los negocios."},
		"Waylon":     {"¡Sí, señor!", "¡Siempre a la orden!", "¡Con gusto, señor!"},
		"Seymour":    {"¡Por el amor de Dios!", "¿Acaso no es maravilloso odiar las mismas cosas?", "¡No toleraré más desorden!"},
	}

	// Creamos un generador local para números aleatorios.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Para cada usuario creado, se crean 10 tweets usando las frases correspondientes a su personaje.
	for _, userSeed := range createdUsers {
		// Buscar las frases asignadas al personaje (usando el FirstName).
		phrases, found := characterPhrases[userSeed.Person.FirstName]
		if !found || len(phrases) == 0 {
			// Si no se encuentra, se utiliza una frase por defecto.
			phrases = []string{"¡Viva Los Simpsons!"}
		}

		for i := 1; i <= 10; i++ {
			// Elegir una frase al azar para el personaje usando el generador local.
			phrase := phrases[r.Intn(len(phrases))]
			// Crear el contenido del tweet.
			tweetContent := fmt.Sprintf("Tweet %d de %s %s: %s", i, userSeed.Person.FirstName, userSeed.Person.LastName, phrase)
			// Crear el tweet en el dominio.
			newTweet := &tweetDomain.Tweet{
				UserID:  userSeed.CreatedUserID,
				Content: tweetContent,
			}
			tweetID, err := tc.CreateTweet(ctx, newTweet)
			if err != nil {
				log.Printf("Error creating tweet for user %s: %v", userSeed.CreatedUserID, err)
			} else {
				log.Printf("Tweet created for user %s with tweet ID: %s", userSeed.CreatedUserID, tweetID)
			}
		}
	}

	log.Println("All test data uploaded successfully")
	return nil
}
