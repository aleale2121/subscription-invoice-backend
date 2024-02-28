package handlers

import (
	"net/http"
	"strconv"
	"subscription-service/internal/storage/db"
	"time"

	"github.com/go-chi/chi/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SubscriptionHandler struct {
	SubscriptionPersistence db.SubscriptionPersistence
	Rabbit                  *amqp.Connection
}

func NewSubscriptionHandler(SubscriptionPersistence db.SubscriptionPersistence,
	Rabbit *amqp.Connection) SubscriptionHandler {
	return SubscriptionHandler{
		SubscriptionPersistence: SubscriptionPersistence,
		Rabbit:                  Rabbit,
	}
}

type SubscriptionJSONPayload struct {
	UserID                int       `json:"UserId"`
	ContractStartDate     time.Time `json:"ContractStartDate"`
	Duration              int32     `json:"Duration"`
	DurationUnits         string    `json:"DurationUnits"`
	BillingFrequency      int32     `json:"BillingFrequency"`
	BillingFrequencyUnits string    `json:"BillingFrequencyUnits"`
	Price                 float32   `json:"Price"`
	Currency              string    `json:"Currency"`
	ProductCode           string    `json:"ProductCode"`
	PlanID                string    `json:"PlanID"`
}

func (app *SubscriptionHandler) GetAllSubsciptionHandler(w http.ResponseWriter, r *http.Request) {

	subscriptions, err := app.SubscriptionPersistence.GetAll()
	if err != nil {
		errorJSON(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "User Subscriptions",
		Data:    subscriptions,
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (app *SubscriptionHandler) GetSubsciptionsToBillTodayHandler(w http.ResponseWriter, r *http.Request) {

	subscriptions, err := app.SubscriptionPersistence.GetSubscriptionsToBillToday()
	if err != nil {
		errorJSON(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "Subscriptions Billed Today",
		Data:    subscriptions,
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (app *SubscriptionHandler) GetSubsciptionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse user ID from request params
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	subscription, err := app.SubscriptionPersistence.GetByUserID(userID)
	if err != nil {
		errorJSON(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "User Subscription",
		Data:    subscription,
	}

	writeJSON(w, http.StatusAccepted, resp)
}
