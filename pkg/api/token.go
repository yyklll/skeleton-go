package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"io/ioutil"

	"github.com/dgrijalva/jwt-go"
	"github.com/yyklll/skeleton/pkg/log"
)

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
func (s *Server) tokenGenerateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("reading token request body failed", err)
		ErrorResponseWithCode(w, r, "invalid request body", http.StatusBadRequest)
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
	t, err := token.SignedString([]byte(s.config.SrvConfig.JWTSecret))
	if err != nil {
		ErrorResponseWithCode(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	var result = TokenResponse{
		Token:     t,
		ExpiresAt: time.Unix(claims.StandardClaims.ExpiresAt, 0),
	}

	JSONResponse(w, r, result)
}

// TokenValidate godoc
// @Summary Validate JWT token
// @Description validates the JWT token
// @Tags HTTP API
// @Accept json
// @Produce json
// @Router /token/validate [post]
// @Success 200 {object} api.TokenValidationResponse
// @Failure 401 {string} string "Unauthorized"
// Get: JWT=$(curl -s -d 'test' localhost:9898/token | jq -r .token)
// Post: curl -H "Authorization: Bearer ${JWT}" hostname:port/token/validate
func (s *Server) tokenValidateHandler(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("authorization")
	if authorizationHeader == "" {
		ErrorResponseWithCode(w, r, "authorization bearer header required", http.StatusUnauthorized)
		return
	}
	bearerToken := strings.Split(authorizationHeader, " ")
	if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
		ErrorResponseWithCode(w, r, "authorization bearer header required", http.StatusUnauthorized)
		return
	}

	claims := jwtCustomClaims{}
	token, err := jwt.ParseWithClaims(bearerToken[1], &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(s.config.SrvConfig.JWTSecret), nil
	})
	if err != nil {
		ErrorResponseWithCode(w, r, err.Error(), http.StatusUnauthorized)
		return
	}

	if token.Valid {
		if claims.StandardClaims.Issuer != "skeleton" {
			ErrorResponseWithCode(w, r, "invalid issuer", http.StatusUnauthorized)
		} else {
			var result = TokenValidationResponse{
				TokenName: claims.Name,
				ExpiresAt: time.Unix(claims.StandardClaims.ExpiresAt, 0),
			}
			JSONResponse(w, r, result)
		}
	} else {
		ErrorResponseWithCode(w, r, "Invalid authorization token", http.StatusUnauthorized)
	}
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TokenValidationResponse struct {
	TokenName string    `json:"token_name"`
	ExpiresAt time.Time `json:"expires_at"`
}
