package authe

// import (
// 	"fmt"
// 	"net/http"

// 	oauth2 "github.com/devpablocristo/monorepo/pkg/authe/oauth2"
// )

// type oAuth2Service struct {
// 	oa oauth2.Service
// }

// func NewOAuth2Service() (OAuth2Service, error) {

// 	oa, err := oauth2.Bootstrap("", "", "", "", "", nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize OAuth2 service: %w", err)
// 	}

// 	return oAuth2Service{
// 		oa: oa,
// 	}, nil
// }

// func (oa oAuth2Service) LoginRedirect(w http.ResponseWriter, r *http.Request, svc oauth2.Service) {
// 	state := oauth2.GenerateState()
// 	// Guardar 'state' en sesión o cache para validarlo luego
// 	// session.Set(r, "oauth2_state", state)

// 	authURL := svc.GetAuthCodeURL(state)
// 	http.Redirect(w, r, authURL, http.StatusFound)
// }

// func (oa oAuth2Service) oAuth2Callback(w http.ResponseWriter, r *http.Request, svc oauth2.Service) {
// 	// Leer code y state de la URL
// 	code := r.URL.Query().Get("code")
// 	state := r.URL.Query().Get("state")

// 	_ = state

// 	// Validar 'state' con el valor guardado en sesión
// 	// savedState := session.Get(r, "oauth2_state")
// 	// if state != savedState {
// 	//     http.Error(w, "invalid state parameter", http.StatusBadRequest)
// 	//     return
// 	// }

// 	// Intercambiar el code por un token
// 	token, err := svc.ExchangeCode(r.Context(), code)
// 	if err != nil {
// 		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
// 		return
// 	}

// 	// Guardar el token en tu DB, cache, cookies seguras, etc.

// 	fmt.Fprintf(w, "Access Token: %s\n", token.AccessToken)
// }
