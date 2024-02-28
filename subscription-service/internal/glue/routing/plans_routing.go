package routing

import (
	"net/http"

	h "subscription-service/internal/handlers"
	"subscription-service/platforms/routers"
)

func PlansRouting(handler *h.PlanHandler) []routers.Route {
	return []routers.Route{
		{
			Method:      http.MethodPost,
			Path:        "/plans",
			Handle:      handler.CreatePlan,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodGet,
			Path:        "/plans",
			Handle:      handler.GetAllPlans,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodGet,
			Path:        "/plans/{id}",
			Handle:      handler.GetPlanByID,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodPut,
			Path:        "/plans/{id}",
			Handle:      handler.UpdatePlan,
			MiddleWares: []http.HandlerFunc{},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/plans/{id}",
			Handle:      handler.DeletePlan,
			MiddleWares: []http.HandlerFunc{},
		},
	}
}
