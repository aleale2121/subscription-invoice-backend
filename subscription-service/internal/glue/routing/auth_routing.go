package routing

import (
	"net/http"

	h "subscription-service/internal/handlers"
	"subscription-service/platforms/routers"
)

func AuthRouting(handler h.AuthHandler) []routers.Route {
	return []routers.Route{
		{
			Method:      http.MethodPost,
			Path:        "/login",
			Handle:      handler.LoginHandler,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodPost,
			Path:        "/signup",
			Handle:      handler.SignUpHandler,
			MiddleWares: []http.HandlerFunc{},
		},
	}
}
