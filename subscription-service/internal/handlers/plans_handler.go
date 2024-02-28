package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"github.com/go-chi/chi/v5"

	"subscription-service/internal/storage/db"

)

type PlanHandler struct {
	PlanPersistence db.PlanPersistence
}

func NewPlanHandler(planPersistence db.PlanPersistence) *PlanHandler {
	return &PlanHandler{
		PlanPersistence: planPersistence,
	}
}

type PlanJSONPayload struct {
	Name                  string  `json:"Name"`
	Duration              int32   `json:"Duration"`
	DurationUnits         string  `json:"DurationUnits"`
	BillingFrequency      int32   `json:"BillingFrequency"`
	BillingFrequencyUnits string  `json:"BillingFrequencyUnits"`
	Price                 float32 `json:"Price"`
	Currency              string  `json:"Currency"`
}

func (h *PlanHandler) GetAllPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.PlanPersistence.GetAllPlans()
	if err != nil {
		errorJSON(w, errors.New("failed to fetch plans"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Subscription Plans",
		Data:    plans,
	}

	writeJSON(w, http.StatusAccepted, payload)
}

func (h *PlanHandler) GetPlanByID(w http.ResponseWriter, r *http.Request) {
	// Parse plan ID from request params
	planID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	plan, err := h.PlanPersistence.GetPlanByID(planID)
	if err != nil {
		errorJSON(w, errors.New("failed to fetch plan"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Plan",
		Data:    plan,
	}

	writeJSON(w, http.StatusAccepted, payload)
}

func (h *PlanHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	// Parse plan from request body
	var requestPayload PlanJSONPayload
	err := readJSON(w, r, &requestPayload)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	planID, err := h.PlanPersistence.InsertPlan(db.Plan{
		Name:                  requestPayload.Name,
		Duration:              requestPayload.Duration,
		DurationUnits:         requestPayload.DurationUnits,
		BillingFrequency:      requestPayload.BillingFrequency,
		BillingFrequencyUnits: requestPayload.BillingFrequencyUnits,
		Price:                 requestPayload.Price,
		Currency:              requestPayload.Currency,
	})
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Plan Created",
		Data:    planID,
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *PlanHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	// Parse plan ID from request params
	planID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Parse plan from request body
	var requestPayload PlanJSONPayload
	err = readJSON(w, r, &requestPayload)
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if err := h.PlanPersistence.UpdatePlan(db.Plan{
		ID:                    planID,
		Name:                  requestPayload.Name,
		Duration:              requestPayload.Duration,
		DurationUnits:         requestPayload.DurationUnits,
		BillingFrequency:      requestPayload.BillingFrequency,
		BillingFrequencyUnits: requestPayload.BillingFrequencyUnits,
		Price:                 requestPayload.Price,
	}); err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Plan Updated",
		Data:    planID,
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *PlanHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	// Parse plan ID from request params
	planID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if err := h.PlanPersistence.DeletePlan(planID); err != nil {
		errorJSON(w, errors.New("failed to delete plan"), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
