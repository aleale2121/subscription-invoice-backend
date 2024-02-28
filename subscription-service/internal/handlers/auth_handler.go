package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"subscription-service/internal/storage/db"
	event "subscription-service/platforms/message-broker/rabbitmq"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AuthHandler struct {
	AuthPersistence         db.UserPersistence
	PlanPersistence         db.PlanPersistence
	SubscriptionPersistence db.SubscriptionPersistence
	Rabbit                  *amqp.Connection
}

// InvoicePayload is the embedded type (in RequestPayload) that describes a request to process invoice
type InvoicePayload struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

const (
	PENDING  = "PENDING-PAYMENT"
	ACTIVE   = "ACTIVE"
	INACTIVE = "INACTIVE"
)

func NewAuthHandler(AuthPersistence db.UserPersistence, PlanPersistence db.PlanPersistence,
	SubscriptionPersistence db.SubscriptionPersistence, Rabbit *amqp.Connection) AuthHandler {
	return AuthHandler{
		AuthPersistence:         AuthPersistence,
		PlanPersistence:         PlanPersistence,
		SubscriptionPersistence: SubscriptionPersistence,
		Rabbit:                  Rabbit,
	}
}
func (app *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := readJSON(w, r, &requestPayload)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against the database
	user, err := app.AuthPersistence.GetByEmail(requestPayload.Email)
	if err != nil {
		errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	writeJSON(w, http.StatusAccepted, payload)
}

func (app *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email             string `json:"Email"`
		Password          string `json:"Password"`
		FirstName         string `json:"FirstName"`
		LastName          string `json:"LastName"`
		PlanID            int    `json:"PlanID"`
		ContractStartDate string `json:"ContractStartDate"`
		ProductCode       string `json:"ProductCode"`
	}

	err := readJSON(w, r, &requestPayload)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	contractStartDate, err := time.Parse("2006-01-02", requestPayload.ContractStartDate)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if contractStartDate.Day()-time.Now().Day() < 0 {
		errorJSON(w, errors.New("invalid contract start date"), http.StatusBadRequest)
		return
	}

	// validate the user against the database
	user, _ := app.AuthPersistence.GetByEmail(requestPayload.Email)
	if user != nil {
		errorJSON(w, errors.New("email already exist"), http.StatusBadRequest)
		return
	}

	// validate the subscription plan against the database
	plan, err := app.PlanPersistence.GetPlanByID(requestPayload.PlanID)
	if err != nil {
		errorJSON(w, errors.New("invalid plan id"), http.StatusBadRequest)
		return
	}

	userID, err := app.AuthPersistence.Insert(db.User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    1,
	})

	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	_, err = app.SubscriptionPersistence.Insert(db.Subscription{
		ID:                    userID,
		UserID:                userID,
		PlanID:                plan.ID,
		ContractStartDate:     contractStartDate,
		Duration:              plan.Duration,
		DurationUnits:         plan.DurationUnits,
		BillingFrequency:      plan.BillingFrequency,
		BillingFrequencyUnits: plan.BillingFrequencyUnits,
		Price:                 plan.Price,
		Currency:              plan.Currency,
		ProductCode:           requestPayload.ProductCode,
		Status:                ACTIVE,
		NextBillingDate:       contractStartDate,
	})
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}
	app.pushToQueue("invoice", userID)

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Signup user %s", requestPayload.Email),
		Data:    userID,
	}

	writeJSON(w, http.StatusAccepted, payload)
}

// pushToQueue pushes a message into RabbitMQ
func (app *AuthHandler) pushToQueue(name string, data any) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := InvoicePayload{
		Name: name,
		Data: data,
	}

	j, err := json.MarshalIndent(&payload, "", "\t")
	if err != nil {
		return err
	}

	err = emitter.Push(string(j), "invoice.SEND")
	if err != nil {
		return err
	}
	return nil
}
