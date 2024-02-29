package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"subscription-service/internal/constants/models"
	"subscription-service/internal/glue/routing"
	"subscription-service/internal/handlers"
	event "subscription-service/internal/message-broker/rabbitmq"
	"subscription-service/internal/pkg/mailer"
	"subscription-service/internal/pkg/scheduler"
	"subscription-service/internal/storage/db"
	"subscription-service/platforms/routers"
	"time"

	ig "subscription-service/internal/pkg/invoicegenerator"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"
)

const webPort = "80"

var counts int64

type Config struct {
	DB     *sql.DB
	Rabbit *amqp.Connection
}

func main() {
	log.Println("Starting user service")

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	userPersistence := db.NewUsersPersistence(conn)
	subcPersistence := db.NewSubscriptionsPersistence(conn)
	planPersistence := db.NewPlansPersistence(conn)

	authHandler := handlers.NewAuthHandler(userPersistence, planPersistence, subcPersistence, rabbitConn)
	subcHandler := handlers.NewSubscriptionHandler(subcPersistence, rabbitConn)
	planHandler := handlers.NewPlanHandler(planPersistence)

	authRouting := routing.AuthRouting(authHandler)
	planRouting := routing.PlansRouting(planHandler)
	subcRouring := routing.SubscriptionRouting(subcHandler)

	var routesList []routers.Route
	routesList = append(routesList, authRouting...)
	routesList = append(routesList, subcRouring...)
	routesList = append(routesList, planRouting...)

	wait := make(chan bool)
	cronJobRunner := cron.New()
	schedulerService := scheduler.NewSchedulerService(cronJobRunner, rabbitConn)
	go schedulerService.Schedules(wait)

	companyAddress := models.Address{
		Address:    "Steinstraße 2",
		Address2:   "Düsseldorf",
		PostalCode: "40212",
		City:       "Stadtbezirk",
		Country:    "Germany",
	}
	invoiceGenerator := ig.NewInvoiceGenerator(
		"Movido",
		"ref",
		"1.0",
		companyAddress)

	mail := createMail()
	// create consumer
	consumer, err := event.NewConsumer(rabbitConn, invoiceGenerator, mail)
	if err != nil {
		// start listening for messages
		log.Println("Listening for and consuming RabbitMQ messages...")
		panic(err)
	}

	// watch the queue and consume events
	go func(eventConsumer event.Consumer) {
		err = eventConsumer.Listen([]string{"invoice.SEND"})
		if err != nil {
			log.Println(err)
		}
	}(consumer)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: routers.Routes(routesList),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

	<-wait

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}

func createMail() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
	}

	return m
}
