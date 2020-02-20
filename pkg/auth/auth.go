package auth

import (
	"fmt"
	"net/http"
	"time"
)

var driverFactories = make(map[string]AuthenticationDriverFactory)

func Register(name string, factory AuthenticationDriverFactory) {
	if factory == nil {
		panic("Must not provide nil AuthenticationDriverFactory")
	}
	_, registered := driverFactories[name]
	if registered {
		panic(fmt.Sprintf("AuthenticationDriverFactory named %s already registered", name))
	}
	driverFactories[name] = factory
}

func Create(name string, parameters map[string]interface{}) (AuthenticationDriver, error) {
	driverFactory, ok := driverFactories[name]
	if !ok {
		return nil, fmt.Errorf("Invalid AuthenticationDriverFactory name %s", name)
	}
	return driverFactory.Create(parameters)
}

type AuthenticationDriverFactory interface {
	Create(parameters map[string]interface{}) (AuthenticationDriver, error)
}

type AuthenticationDriver interface {
	// HTTP handler generating token
	TokenGenerateHandler(w http.ResponseWriter, r *http.Request)

	// HTTP middleware validating token
	TokenValidateMiddleware(next http.Handler) http.Handler
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
