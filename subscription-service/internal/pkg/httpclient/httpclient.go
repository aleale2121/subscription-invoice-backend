package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var url = "https://jsonplaceholder.typicode.com/todos/"

func PostInvoiceToAccountingService(invoiceID string) {

	jsonData, err := json.Marshal(Todo{
		UserID:    1,
		ID:        1,
		Title:     invoiceID,
		Completed: true,
	})
	if err != nil {
		log.Printf("Err PostInvoiceToAccountingService: +%v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Err PostInvoiceToAccountingService: +%v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Err PostInvoiceToAccountingService: +%v", err)
		return
	}

	fmt.Println("POST request successful")
}

type Todo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}
