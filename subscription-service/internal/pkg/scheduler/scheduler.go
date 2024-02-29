package scheduler

import (
	"fmt"
	"log"
	"subscription-service/internal/constants/models"
	"subscription-service/internal/constants/states"
	ig "subscription-service/internal/pkg/invoicegenerator"
	"subscription-service/internal/pkg/mailer"
	"subscription-service/internal/storage/db"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron                    *cron.Cron
	Rabbit                  *amqp.Connection
	subscriptionPersistence db.SubscriptionPersistence
	planPersistence         db.PlanPersistence
	userPersistence         db.UserPersistence
	invoicePersistence      db.InvoicePersistence
	invoicegenerator        ig.InvoiceGenerator
	mail                    mailer.Mail
}

func NewSchedulerService(cron *cron.Cron, Rabbit *amqp.Connection,
	subscriptionPersistence db.SubscriptionPersistence,
	planPersistence db.PlanPersistence,
	userPersistence db.UserPersistence,
	invoicePersistence db.InvoicePersistence,
	invoicegenerator ig.InvoiceGenerator,
	mail mailer.Mail,
) SchedulerService {
	return SchedulerService{
		cron:                    cron,
		Rabbit:                  Rabbit,
		subscriptionPersistence: subscriptionPersistence,
		planPersistence:         planPersistence,
		userPersistence:         userPersistence,
		invoicePersistence:      invoicePersistence,
		invoicegenerator:        invoicegenerator,
		mail:                    mail,
	}
}

func (s *SchedulerService) Schedules(wait chan bool) {
	_, err := s.cron.AddFunc("@daily", s.ProcessInvoices)
	if err != nil {
		log.Printf("Error: scheduling daily invoice processing %v", err.Error())
		return
	}
	_, err = s.cron.AddFunc("@daily", s.ProcessFailedInvoices)
	if err != nil {
		log.Printf("Error: scheduling failed invoice processing %v", err.Error())
		return
	}
	s.cron.Run()
}

func (s *SchedulerService) ProcessInvoices() {
	subscriptions, err := s.subscriptionPersistence.GetSubscriptionsToBillToday()
	if err != nil {
		log.Printf("Error: fetching subscription %v", err.Error())
		return
	}
	for _, subscription := range subscriptions {
		go func(subscription *models.Subscription) {
			plan, err := s.planPersistence.GetPlanByID(subscription.PlanID)
			if err != nil {
				log.Printf("Error: fetching plan %v", err.Error())
				return
			}
			user, err := s.userPersistence.GetOne(subscription.UserID)
			if err != nil {
				log.Printf("Error: fetching user %v", err.Error())
				return
			}
			addr, err := s.userPersistence.GetBillingAddressByUserID(subscription.UserID)
			if err != nil {
				log.Printf("Error: fetching subscription %v", err.Error())
				return
			}

			invoiceID, err := sendInvoice(models.InvoicePayload{
				User:           *user,
				BillingAddress: *addr,
				Subscription:   *subscription,
				Plan:           *plan,
			}, "", s.invoicegenerator, s.mail)
			if err != nil {
				log.Printf("Error: sending invocie %v", err.Error())
				s.invoicePersistence.AddFailedInvoice(models.FailedInvoice{
					SubscriptionID: subscription.ID,
					InvoiceID:      invoiceID,
					InvoiceDate:    time.Now(),
					EmailRetry:     1,
				})
				return
			}
		}(subscription)
	}
}

func (s *SchedulerService) ProcessFailedInvoices() {
	failedInvoices, err := s.invoicePersistence.GetAllInvoices()
	if err != nil {
		log.Printf("Error: fetching failed invoices %v", err.Error())
		return
	}
	for _, invoice := range failedInvoices {
		go func(invoice models.FailedInvoice) {
			subscription, err := s.subscriptionPersistence.GetOne(invoice.SubscriptionID)
			if err != nil {
				log.Printf("Error: fetching plan %v", err.Error())
				return
			}
			plan, err := s.planPersistence.GetPlanByID(subscription.PlanID)
			if err != nil {
				log.Printf("Error: fetching plan %v", err.Error())
				return
			}
			user, err := s.userPersistence.GetOne(subscription.UserID)
			if err != nil {
				log.Printf("Error: fetching user %v", err.Error())
				return
			}
			addr, err := s.userPersistence.GetBillingAddressByUserID(subscription.UserID)
			if err != nil {
				log.Printf("Error: fetching subscription %v", err.Error())
				return
			}

			invoiceID, err := sendInvoice(models.InvoicePayload{
				User:           *user,
				BillingAddress: *addr,
				Subscription:   *subscription,
				Plan:           *plan,
			}, invoice.InvoiceID, s.invoicegenerator, s.mail)
			if err != nil {
				log.Printf("Error: re-sending invocie %v", err.Error())
				s.invoicePersistence.UpdateInvoice(invoiceID, invoice.EmailRetry+1)
				return
			} else {
				s.invoicePersistence.DeleteInvoice(invoiceID)
			}
		}(invoice)
	}
}

func sendInvoice(invoicePld models.InvoicePayload, invoiceID string, invoicegenerator ig.InvoiceGenerator, mail mailer.Mail) (string, error) {
	if invoiceID == "" {
		// Generate an Invoice
		invoiceItem := models.InvoiceItem{
			Name:        fmt.Sprintf("%s %s", invoicePld.Subscription.ProductCode, invoicePld.Plan.Name),
			Description: "",
			UnitCost:    fmt.Sprint(invoicePld.Subscription.Price / float32(invoicePld.Subscription.BillingFrequency)),
			Quantity:    "1",
			Tax:         "0",
			Discount:    "0",
		}

		invoice := models.Invoice{
			Date:         time.Now().Format(states.TIME_LAYOUT),
			PaymentTerm:  time.Now().Add(time.Hour * 24 * 7).Format(states.TIME_LAYOUT),
			CustomerName: fmt.Sprintf("%s %s", invoicePld.User.FirstName, invoicePld.User.LastName),
			Description:  "Invoice Description",
			Notes:        "",
			CustomerAddress: models.Address{
				Address:    invoicePld.BillingAddress.Address,
				Address2:   invoicePld.BillingAddress.Address2,
				PostalCode: invoicePld.BillingAddress.PostalCode,
				City:       invoicePld.BillingAddress.City,
				Country:    invoicePld.BillingAddress.Country,
			},
			Items: []models.InvoiceItem{invoiceItem},
		}
		id, err := invoicegenerator.Generate(invoice)
		if err != nil {
			log.Println("Error: generate invoice")
			log.Println(err)
			return "", err
		}
		invoiceID = id
	}

	// Send Email
	invoicePDF := fmt.Sprintf("../../../temp/%s.pdf", invoiceID)
	log.Printf("Mail: %+v", mail)

	err := mail.SendSMTPMessage(mailer.Message{
		To:          invoicePld.User.Email,
		Subject:     "Invoice For Movido Subscription",
		Attachments: []string{invoicePDF},
		Data:        "Invoice",
		DataMap:     map[string]any{"inv": "Invoice"},
	})
	if err != nil {
		log.Println("Error: sending main")
		log.Println(err)
		return invoiceID, err
	}
	return invoiceID, nil
}
