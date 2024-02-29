package invoicegenerator

import (
	"fmt"
	"log"
	"os"
	"subscription-service/internal/constants/models"

	uuid "github.com/satori/go.uuid"

	generator "github.com/angelodlfrtr/go-invoice-generator"
)

type InvoiceGenerator struct {
	CompanyName    string
	Reference      string
	Version        string
	CompanyAddress models.Address
}

func NewInvoiceGenerator(CompanyName, Reference, Version string, CompanyAddress models.Address) InvoiceGenerator {
	return InvoiceGenerator{
		CompanyName:    CompanyName,
		Reference:      Reference,
		Version:        Version,
		CompanyAddress: CompanyAddress,
	}
}
func (ig InvoiceGenerator) Generate(invoice models.Invoice) (string, error) {
	doc, _ := generator.New(generator.Invoice, &generator.Options{
		TextTypeInvoice: "INVOICE",
		AutoPrint:       true,
	})

	doc.SetHeader(&generator.HeaderFooter{
		Text:       fmt.Sprintf("<center>Thank you for choosing %s</center>", ig.CompanyName),
		Pagination: true,
	})

	doc.SetFooter(&generator.HeaderFooter{
		Text:       "<center>Thank you for your prompt payment.</center>",
		Pagination: true,
	})

	doc.SetRef(ig.Reference)
	doc.SetVersion(ig.Version)

	doc.SetDescription(invoice.Description)
	doc.SetNotes(invoice.Notes)

	doc.SetDate(invoice.Date)
	doc.SetPaymentTerm(invoice.PaymentTerm)

	logoBytes, err := os.ReadFile("./Movido_Logo.png")
	if err != nil {
		log.Fatal(err)
	}

	doc.SetCompany(&generator.Contact{
		Name: ig.CompanyName,
		Logo: logoBytes,
		Address: &generator.Address{
			Address:    ig.CompanyAddress.Address,
			Address2:   ig.CompanyAddress.Address2,
			PostalCode: ig.CompanyAddress.PostalCode,
			City:       ig.CompanyAddress.City,
			Country:    ig.CompanyAddress.Country,
		},
	})

	doc.SetCustomer(&generator.Contact{
		Name: invoice.CustomerName,
		Address: &generator.Address{
			Address:    invoice.CustomerAddress.Address,
			PostalCode: invoice.CustomerAddress.PostalCode,
			City:       invoice.CustomerAddress.City,
			Country:    invoice.CustomerAddress.Country,
		},
	})

	for _, item := range invoice.Items {
		doc.AppendItem(&generator.Item{
			Name:        item.Name,
			Description: item.Description,
			UnitCost:    item.UnitCost,
			Quantity:    item.Quantity,
			Tax: &generator.Tax{
				Percent: item.Tax,
			},
		})
	}

	pdf, err := doc.Build()
	if err != nil {
		log.Fatal(err)
	}

	id := uuid.NewV4().String()
	err = pdf.OutputFileAndClose(fmt.Sprintf("../../temp/%s.pdf", id))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return id, nil
}
