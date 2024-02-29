package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"subscription-service/internal/constants/models"
	"subscription-service/internal/constants/states"

	event "subscription-service/internal/message-broker/rabbitmq"
	"subscription-service/internal/storage/db"
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
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Address struct {
	Address    string `json:"address" validate:"required"`
	Address2   string `json:"address_2,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
}

type SignupRequestPayload struct {
	Email             string  `json:"Email"`
	Password          string  `json:"Password"`
	FirstName         string  `json:"FirstName"`
	LastName          string  `json:"LastName"`
	PlanID            int     `json:"PlanID"`
	ContractStartDate string  `json:"ContractStartDate"`
	ProductCode       string  `json:"ProductCode"`
	Address           Address `json:"Address"`
}

func NewAuthHandler(AuthPersistence db.UserPersistence, PlanPersistence db.PlanPersistence,
	SubscriptionPersistence db.SubscriptionPersistence, Rabbit *amqp.Connection) AuthHandler {
	return AuthHandler{
		AuthPersistence:         AuthPersistence,
		PlanPersistence:         PlanPersistence,
		SubscriptionPersistence: SubscriptionPersistence,
		Rabbit:                  Rabbit,
	}
}

func (app *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {

	var requestPayload SignupRequestPayload

	err := readJSON(w, r, &requestPayload)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	contractStartDate, err := time.Parse(states.TIME_LAYOUT, requestPayload.ContractStartDate)
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

	usr := models.User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    1,
	}

	userID, err := app.AuthPersistence.AddUser(usr)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}
	usr.ID = userID

	addr := models.Address{
		UserID:     userID,
		Address:    requestPayload.Address.Address,
		Address2:   requestPayload.Address.Address2,
		PostalCode: requestPayload.Address.PostalCode,
		City:       requestPayload.Address.City,
		Country:    requestPayload.Address.Country,
	}

	addrID, err := app.AuthPersistence.AddBillingAddress(addr)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}
	addr.ID = addrID

	subs := models.Subscription{
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
		Status:                states.ACTIVE,
		NextBillingDate:       contractStartDate,
	}

	subsID, err := app.SubscriptionPersistence.AddSubscription(subs)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}
	subs.ID = subsID
	now, _ := time.Parse(states.TIME_LAYOUT, time.Now().Format(states.TIME_LAYOUT))
	
	if now.Equal(contractStartDate) {
		app.pushToQueue("invoice", models.InvoicePayload{
			User:           usr,
			BillingAddress: addr,
			Subscription:   subs,
			Plan:           *plan,
		})
	}

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

	j, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	err = emitter.Push(string(j), "invoice.SEND")
	if err != nil {
		return err
	}
	return nil
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
