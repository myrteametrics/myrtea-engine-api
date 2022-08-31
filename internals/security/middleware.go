package security

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"

	"go.uber.org/zap"
)

// Middleware is an interface for standard http middleware
type Middleware interface {
	Handler(h http.Handler) http.Handler
}

//
// TODO: Must be refactored and merged in a single lib with github.com/auth0/go-jwt-middleware/jwtmiddleware.go (instead of wrapping it)
//

// MiddlewareJWT is an implementation of Middleware interface, which provides a specific security handler based on JWT (JSON Web Token)
type MiddlewareJWT struct {
	Auth       Auth
	Handler    func(h http.Handler) http.Handler
	signingKey []byte
}

// JwtToken wrap the json web token string
type JwtToken struct {
	Token string `json:"token"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandStringWithCharset generate a random string with a specific charset
func RandStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandString generate a random string with the default charset ([A-Za-z])
func RandString(length int) string {
	return RandStringWithCharset(length, charset)
}

// NewMiddlewareJWT initialize a new instance of MiddlewareJWT and returns a pointer of it
func NewMiddlewareJWT(jwtSigningKey []byte, auth Auth) *MiddlewareJWT {
	/*
		var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return jwtSigningKey, nil
			},
			SigningMethod: jwt.SigningMethodHS256,
			Debug:         true,
		})
		securingHandler := jwtMiddleware.Handler
	*/
	return &MiddlewareJWT{auth, nil, jwtSigningKey}
}

// GetToken returns a http.Handler to authenticate and get a JWT
func (middleware *MiddlewareJWT) GetToken() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var credentials users.UserWithPassword
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			zap.L().Error("GetToken.Decode:", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, allowed, err := middleware.Auth.Authenticate(credentials.Login, credentials.Password)
		if err != nil {
			zap.L().Warn("Authentication failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !allowed {
			zap.L().Error("Invalid credentials")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss": "Myrtea metrics",
			"exp": time.Now().Add(time.Hour * 12).Unix(),
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"id":  user.ID,
		})

		// Sign the token with our signing key
		tokenString, err := token.SignedString(middleware.signingKey)
		if err != nil {
			zap.L().Error("Error while signing token", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
		if err != nil {
			zap.L().Error("Error while encoding the token ", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	})
}

// AdminAuthentificator is a middle which check if the user is administrator (role=1)
func AdminAuthentificator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get user infos from token
		_, claims, _ := jwtauth.FromContext(r.Context())
		// test if user haven't right to access
		if claims["role"] != float64(1) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handler return a specific securing Handler. It should be used to wrap other handlers which must be secured with JWT
/*
func (middleware *MiddlewareJWT) Handler(next http.Handler) http.Handler {
	return middleware.Handler(next)
}
*/
