package event

import (
	"encoding/json"
	"fmt"
	"log"
	"subscription-service/internal/constants/models"
	"subscription-service/internal/constants/states"
	ig "subscription-service/internal/pkg/invoicegenerator"
	"subscription-service/internal/pkg/mailer"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn             *amqp.Connection
	invoicegenerator ig.InvoiceGenerator
	mail             mailer.Mail
}

func NewConsumer(conn *amqp.Connection, invoicegenerator ig.InvoiceGenerator, mail mailer.Mail) (Consumer, error) {
	consumer := Consumer{
		conn:             conn,
		invoicegenerator: invoicegenerator,
		mail:             mail,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,
			s,
			"invoice_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			switch payload.Name {
			case "invoice":
				go sendInvoice(payload, consumer.invoicegenerator, consumer.mail)
			default:
				log.Printf("recieved via rabbit %+v \n", payload)
			}
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [invoice_topic, %s]\n", q.Name)
	<-forever

	return nil
}

// User is the structure which holds one user from the database.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Active    int       `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

type InvoicePayload struct {
	User           User         `json:"User"`
	BillingAddress Address      `json:"BillingAddress"`
	Subscription   Subscription `json:"Subscription"`
	Plan           Plan         `json:"Plan"`
}

func sendInvoice(payload Payload, invoicegenerator ig.InvoiceGenerator, mail mailer.Mail) {
	// Parse InvoicePayload
	var invoicePld InvoicePayload
	j, err := json.Marshal(payload.Data)
	if err != nil {
		log.Println("Error: marshal")
		log.Println(err)
		return
	}
	err = json.Unmarshal(j, &invoicePld)
	if err != nil {
		fmt.Println("Error: unmarshal")
		fmt.Println(err)
		return
	}
	log.Printf("recieved invoice payload %+v \n", invoicePld)

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
	invoiceID, err := invoicegenerator.Generate(invoice)
	if err != nil {
		log.Println("Error: generate invoice")
		log.Println(err)
		return
	}

	// Send Email
	invoicePDF := fmt.Sprintf("../../../temp/%s.pdf", invoiceID)
	log.Printf("Mail: %+v", mail)

	err = mail.SendSMTPMessage(mailer.Message{
		To:          invoicePld.User.Email,
		Subject:     "Invoice For Movido Subscription",
		Attachments: []string{invoicePDF},
		Data:        "Invoice",
		DataMap:     map[string]any{"inv": "Invoice"},
	})
	if err != nil {
		log.Println("Error: sending main")
		log.Println(err)
	}
}
