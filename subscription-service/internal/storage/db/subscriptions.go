package db

import (
	"context"
	"database/sql"
	"log"
	"subscription-service/internal/constants/models"
	"time"
)

type SubscriptionPersistence struct {
	db *sql.DB
}

// NewSubscriptionsPersistence is the function used to create an instance of the SubscriptionPersistence.
func NewSubscriptionsPersistence(dbPool *sql.DB) SubscriptionPersistence {
	return SubscriptionPersistence{db: dbPool}
}

// GetAll returns a slice of all subscriptions, sorted by ID
func (s *SubscriptionPersistence) GetAll() ([]*models.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT * FROM subscriptions ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription

	for rows.Next() {
		var subscription models.Subscription
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.PlanID,
			&subscription.ContractStartDate,
			&subscription.Duration,
			&subscription.DurationUnits,
			&subscription.BillingFrequency,
			&subscription.BillingFrequencyUnits,
			&subscription.Price,
			&subscription.Currency,
			&subscription.ProductCode,
			&subscription.Status,
			&subscription.BilledCycles,
			&subscription.NextBillingDate,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		subscriptions = append(subscriptions, &subscription)
	}

	return subscriptions, nil
}

// GetByUserID returns one subscription by UserID
func (s *SubscriptionPersistence) GetByUserID(userID int) (*models.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT * FROM subscriptions WHERE user_id = $1`

	var subscription models.Subscription
	row := s.db.QueryRowContext(ctx, query, userID)

	err := row.Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.ContractStartDate,
		&subscription.Duration,
		&subscription.DurationUnits,
		&subscription.BillingFrequency,
		&subscription.BillingFrequencyUnits,
		&subscription.Price,
		&subscription.Currency,
		&subscription.ProductCode,
		&subscription.Status,
		&subscription.BilledCycles,
		&subscription.NextBillingDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

// GetOne returns one subscription by ID
func (s *SubscriptionPersistence) GetOne(id int) (*models.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT * FROM subscriptions WHERE id = $1`

	var subscription models.Subscription
	row := s.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.ContractStartDate,
		&subscription.Duration,
		&subscription.DurationUnits,
		&subscription.BillingFrequency,
		&subscription.BillingFrequencyUnits,
		&subscription.Price,
		&subscription.Currency,
		&subscription.ProductCode,
		&subscription.Status,
		&subscription.BilledCycles,
		&subscription.NextBillingDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

// Update updates one subscription in the database, using the information stored in the receiver sub
func (s *SubscriptionPersistence) Update(subscription models.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE subscriptions SET
        user_id = $1,
        plan_id = $2,
        contract_start_date = $3,
        duration = $4,
        duration_units = $5,
        billing_frequency = $6,
        billing_frequency_units = $7,
        price = $8,
        currency = $9,
        product_code = $10,
        status = $11,
        next_billing_date = $12,
		updated_At = $13,
        WHERE id = $14
    `

	_, err := s.db.ExecContext(ctx, stmt,
		subscription.UserID,
		subscription.PlanID,
		subscription.ContractStartDate,
		subscription.Duration,
		subscription.DurationUnits,
		subscription.BillingFrequency,
		subscription.BillingFrequencyUnits,
		subscription.Price,
		subscription.Currency,
		subscription.ProductCode,
		subscription.Status,
		&subscription.BilledCycles,
		subscription.NextBillingDate,
		time.Now(),
		subscription.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes one subscription from the database, by Subscription.ID
func (s *SubscriptionPersistence) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM subscriptions WHERE id = $1`

	_, err := s.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// AddSubscription inserts a new subscription into the database, and returns the ID of the newly inserted row
func (s *SubscriptionPersistence) AddSubscription(subscription models.Subscription) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	stmt := `INSERT INTO subscriptions (user_id, plan_id, contract_start_date, duration, duration_units, billing_frequency, billing_frequency_units, price, currency, product_code, next_billing_date,status,created_at,updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id
    `

	err := s.db.QueryRowContext(ctx, stmt,
		subscription.UserID,
		subscription.PlanID,
		subscription.ContractStartDate,
		subscription.Duration,
		subscription.DurationUnits,
		subscription.BillingFrequency,
		subscription.BillingFrequencyUnits,
		subscription.Price,
		subscription.Currency,
		subscription.ProductCode,
		subscription.NextBillingDate,
		subscription.Status,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// GetSubscriptionsToBillToday returns subscriptions to be billed today.
func (s *SubscriptionPersistence) GetSubscriptionsToBillToday() ([]*models.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
        SELECT * FROM subscriptions
        WHERE billed_cycles < billing_frequency
        AND status = 'ACTIVE'
        AND next_billing_date::date = $1
    `

	rows, err := s.db.QueryContext(ctx, query, time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription

	for rows.Next() {
		var subscription models.Subscription
		if err := rows.Scan(&subscription.ID,
			&subscription.UserID,
			&subscription.PlanID,
			&subscription.ContractStartDate,
			&subscription.Duration,
			&subscription.DurationUnits,
			&subscription.BillingFrequency,
			&subscription.BillingFrequencyUnits,
			&subscription.Price,
			&subscription.Currency,
			&subscription.ProductCode,
			&subscription.Status,
			&subscription.BilledCycles,
			&subscription.NextBillingDate,
			&subscription.CreatedAt,
			&subscription.UpdatedAt); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &subscription)
	}

	return subscriptions, nil
}
