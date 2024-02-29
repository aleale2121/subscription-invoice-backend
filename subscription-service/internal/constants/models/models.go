package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is the structure which holds one user from the database.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  string    `json:"-"`
	Active    int       `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PasswordMatches uses Go's bcrypt package to compare a user supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// Address represents the structure of a billing address
type Address struct {
	ID         int    `json:"ID"`
	UserID     int    `json:"UserID"`
	Address    string `json:"Address"`
	Address2   string `json:"Address2"`
	PostalCode string `json:"PostalCode"`
	City       string `json:"City"`
	Country    string `json:"Country"`
}

// Plan is the structure which holds one subscription plan from the database.
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

// Subscription is the structure which holds one subscription from the database.
type Subscription struct {
	ID                    int       `json:"ID"`
	UserID                int       `json:"UserID"`
	PlanID                int       `json:"PlanID"`
	ContractStartDate     time.Time `json:"ContractStartDate"`
	Duration              int32     `json:"Duration"`
	DurationUnits         string    `json:"DurationUnits"`
	BillingFrequency      int32     `json:"BillingFrequency"`
	BillingFrequencyUnits string    `json:"BillingFrequencyUnits"`
	Price                 float32   `json:"Price"`
	Currency              string    `json:"Currency"`
	ProductCode           string    `json:"ProductCode"`
	Status                string    `json:"Status"`
	BilledCycles          int       `json:"BilledCycles"`
	NextBillingDate       time.Time `json:"NextBillingDate"`
	CreatedAt             time.Time `json:"CreatedAt"`
	UpdatedAt             time.Time `json:"UpdatedAt"`
}

// FailedInvoice is the structure which holds one failed invoice data.
type FailedInvoice struct {
	ID             int       `json:"ID"`
	SubscriptionID int       `json:"SubscriptionID"`
	InvoiceID      string    `json:"InvoiceID"`
	InvoiceDate    time.Time `json:"InvoiceDate"`
	EmailRetry     int       `json:"EmailRetry"`
}

type InvoicePayload struct {
	User           User         `json:"User"`
	BillingAddress Address      `json:"BillingAddress"`
	Subscription   Subscription `json:"Subscription"`
	Plan           Plan         `json:"Plan"`
}

type Invoice struct {
	Date            string
	PaymentTerm     string
	CustomerName    string
	Description     string
	Notes           string
	CustomerAddress Address
	Items           []InvoiceItem
}

type InvoiceItem struct {
	Name        string
	Description string
	UnitCost    string
	Quantity    string
	Tax         string
	Discount    string
}
