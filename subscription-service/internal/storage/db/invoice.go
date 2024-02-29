package db

import (
	"context"
	"database/sql"
	"log"
	"subscription-service/internal/constants/models"
)

type InvoicePersistence struct {
	db *sql.DB
}

// NewInvoicePersistence is the function used to create an instance of the InvoicePersistence.
func NewInvoicePersistence(dbPool *sql.DB) InvoicePersistence {
	return InvoicePersistence{db: dbPool}
}

// AddFailedInvoice adds a failed invoice record to the database
func (p *InvoicePersistence) AddFailedInvoice(failedInvoice models.FailedInvoice) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "INSERT INTO failed_invoices (subscription_id, invoice_id, invoice_date, email_retry) VALUES ($1, $2, $3, $4) RETURNING id"

	var id int
	err := p.db.QueryRowContext(ctx, stmt, failedInvoice.SubscriptionID, failedInvoice.InvoiceID, failedInvoice.InvoiceDate, failedInvoice.EmailRetry).Scan(&id)
	if err != nil {
		log.Println("Error inserting failed invoice:", err)
		return 0, err
	}
	return id, nil
}

// UpdateInvoice updates an existing invoice record in the database
func (p *InvoicePersistence) UpdateInvoice(invoiceID string, newEmailRetry int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "UPDATE failed_invoices SET email_retry = $1 WHERE invoice_id = $2"

	_, err := p.db.ExecContext(ctx, stmt, newEmailRetry, invoiceID)
	if err != nil {
		log.Println("Error updating invoice:", err)
		return err
	}
	return nil
}

// GetAllInvoices returns all invoice records from the database
func (p *InvoicePersistence) GetAllInvoices() ([]models.FailedInvoice, error) {
	rows, err := p.db.Query("SELECT id, subscription_id, invoice_id, invoice_date, email_retry FROM failed_invoices")
	if err != nil {
		log.Println("Error querying invoices:", err)
		return nil, err
	}
	defer rows.Close()

	var invoices []models.FailedInvoice
	for rows.Next() {
		var invoice models.FailedInvoice
		if err := rows.Scan(&invoice.ID, &invoice.SubscriptionID, &invoice.InvoiceID, &invoice.InvoiceDate, &invoice.EmailRetry); err != nil {
			log.Println("Error scanning invoice row:", err)
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating through invoices:", err)
		return nil, err
	}

	return invoices, nil
}

// DeleteInvoice deletes an invoice record from the database based on the invoice ID
func (p *InvoicePersistence) DeleteInvoice(invoiceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := "DELETE FROM failed_invoices WHERE invoice_id = $1"

	_, err := p.db.ExecContext(ctx, stmt, invoiceID)
	if err != nil {
		log.Println("Error deleting invoice:", err)
		return err
	}
	return nil
}
