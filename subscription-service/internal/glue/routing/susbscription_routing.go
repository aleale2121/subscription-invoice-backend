package routing

import (
	"net/http"

	h "subscription-service/internal/handlers"
	"subscription-service/platforms/routers"
)

func SubscriptionRouting(handler h.SubscriptionHandler) []routers.Route {
	return []routers.Route{
		{
			Method:      http.MethodGet,
			Path:        "/users/subscriptions",
			Handle:      handler.GetAllSubsciptionHandler,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodGet,
			Path:        "/users/subscriptions/todays",
			Handle:      handler.GetSubsciptionsToBillTodayHandler,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodGet,
			Path:        "/users/{id}/subscriptions",
			Handle:      handler.GetSubsciptionHandler,
			MiddleWares: []http.HandlerFunc{},
		},
	}
}
