package db

import (
	"context"
	"database/sql"
	"log"
)

type Plan struct {
	ID                    int     `json:"ID"`
	Name                  string  `json:"Name"`
	Duration              int32   `json:"Duration"`
	DurationUnits         string  `json:"DurationUnits"`
	BillingFrequency      int32   `json:"BillingFrequency"`
	BillingFrequencyUnits string  `json:"BillingFrequencyUnits"`
	Price                 float32 `json:"Price"`
	Currency              string  `json:"Currency"`
}

type PlanPersistence struct {
	db *sql.DB
}

// NewPlansPersistence is the function used to create an instance of the PlanPersistence.
func NewPlansPersistence(dbPool *sql.DB) PlanPersistence {
	return PlanPersistence{db: dbPool}
}

// GetAllPlans returns all plans from the database
func (p *PlanPersistence) GetAllPlans() ([]*Plan, error) {
	rows, err := p.db.Query("SELECT id, name, duration, duration_units, billing_frequency, billing_frequency_units, price, currency FROM plans")
	if err != nil {
		log.Println("Error querying plans:", err)
		return nil, err
	}
	defer rows.Close()

	var plans []*Plan
	for rows.Next() {
		var plan Plan
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.Duration, &plan.DurationUnits, &plan.BillingFrequency, &plan.BillingFrequencyUnits, &plan.Price, &plan.Currency); err != nil {
			log.Println("Error scanning plan row:", err)
			return nil, err
		}
		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating through plans:", err)
		return nil, err
	}

	return plans, nil
}

// GetPlanByID returns a plan from the database by ID
func (p *PlanPersistence) GetPlanByID(id int) (*Plan, error) {
	var plan Plan
	err := p.db.QueryRow("SELECT id, name, duration, duration_units, billing_frequency, billing_frequency_units, price, currency FROM plans WHERE id = $1", id).
		Scan(&plan.ID, &plan.Name, &plan.Duration, &plan.DurationUnits, &plan.BillingFrequency, &plan.BillingFrequencyUnits, &plan.Price, &plan.Currency)
	if err != nil {
		log.Println("Error querying plan by ID:", err)
		return nil, err
	}
	return &plan, nil
}

// UpdatePlan updates a plan in the database
func (p *PlanPersistence) UpdatePlan(plan Plan) error {
	_, err := p.db.Exec("UPDATE plans SET name = $1, duration = $2, duration_units = $3, billing_frequency = $4, billing_frequency_units = $5, price = $6, currency = $7 WHERE id = $8",
		plan.Name, plan.Duration, plan.DurationUnits, plan.BillingFrequency, plan.BillingFrequencyUnits, plan.Price, plan.Currency, plan.ID)
	if err != nil {
		log.Println("Error updating plan:", err)
		return err
	}
	return nil
}

// DeletePlan deletes a plan from the database
func (p *PlanPersistence) DeletePlan(id int) error {
	_, err := p.db.Exec("DELETE FROM plans WHERE id = $1", id)
	if err != nil {
		log.Println("Error deleting plan:", err)
		return err
	}
	return nil
}

// InsertPlan inserts a new plan into the database
func (p *PlanPersistence) InsertPlan(plan Plan) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "INSERT INTO plans (name, duration, duration_units, billing_frequency, billing_frequency_units, price, currency) VALUES ($1, $2, $3, $4, $5, $6, $7) returning id"

	var id int
	err := p.db.QueryRowContext(ctx, stmt, plan.Name, plan.Duration, plan.DurationUnits, plan.BillingFrequency, plan.BillingFrequencyUnits, plan.Price, plan.Currency).Scan(&id)
	if err != nil {
		log.Println("Error inserting plan:", err)
		return 0, err
	}
	return id, nil
}
