package jwt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/yyklll/skeleton/pkg/api"
	"github.com/yyklll/skeleton/pkg/auth"
	"github.com/yyklll/skeleton/pkg/log"
)

func init() {
	auth.Register("jwt", &JWTTokenDriverFactory{})
}

type JWTTokenDriverFactory struct{}

func (jf *JWTTokenDriverFactory) Create(parameters map[string]interface{}) (auth.AuthenticationDriver, error) {
	return &JWTTokenDriver{secret: parameters["secret"].(string)}, nil
}

type JWTTokenDriver struct {
	secret string
}

type jwtCustomClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

// Token godoc
// @Summary Generate JWT token
// @Description issues a JWT token valid for one minute
// @Tags HTTP API
// @Accept json
// @Produce json
// @Router /token [post]
// @Success 200 {object} api.TokenResponse
func (j *JWTTokenDriver) TokenGenerateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("reading token request body failed", err)
		api.ErrorResponseWithCode(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	user := "anonymous"
	if len(body) > 0 {
		user = string(body)
	}

	claims := &jwtCustomClaims{
		user,
		jwt.StandardClaims{
			Issuer:    "skeleton",
			ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secret))
	if err != nil {
		api.ErrorResponseWithCode(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	var result = auth.AuthResponse{
		Token:     t,
		ExpiresAt: time.Unix(claims.StandardClaims.ExpiresAt, 0),
	}

	api.JSONResponse(w, r, result)
}

// TokenValidate godoc
// @Summary Validate JWT token
// @Description validates the JWT token
// @Tags HTTP API
// @Accept json
// @Produce json
// @Success 200 {object} api.TokenValidationResponse
// @Failure 401 {string} string "Unauthorized"
func (j *JWTTokenDriver) TokenValidateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader == "" {
			api.ErrorResponseWithCode(w, r, "authorization bearer header required", http.StatusUnauthorized)
			return
		}
		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			api.ErrorResponseWithCode(w, r, "authorization bearer header required", http.StatusUnauthorized)
			return
		}

		claims := jwtCustomClaims{}
		token, err := jwt.ParseWithClaims(bearerToken[1], &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(j.secret), nil
		})
		if err != nil {
			api.ErrorResponseWithCode(w, r, err.Error(), http.StatusUnauthorized)
			return
		}

		if token.Valid {
			if claims.StandardClaims.Issuer != "skeleton" {
				api.ErrorResponseWithCode(w, r, "invalid issuer", http.StatusUnauthorized)
			} else {
				// TODO: add expiration into header
				next.ServeHTTP(w, r)
			}
		} else {
			api.ErrorResponseWithCode(w, r, "Invalid authorization token", http.StatusUnauthorized)
		}

	})
}
