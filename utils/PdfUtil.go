package utils

import (
	"fmt"
	"time"

	"github.com/fireops-software/fireops-edge-printer/domain"
	appError "github.com/fireops-software/fireops-edge-printer/error"
	"github.com/signintech/gopdf"
)

const (
	xOffset = 40
	yOffset = 20
)

func CreatePdfFromEvents(events []domain.Event) (*gopdf.GoPdf, error) {
	// Create PDF document
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	// Set Font
	err := pdf.AddTTFFont("roboto", "Roboto.ttf")
	if err != nil {
		return nil, appError.NewErrPdfGeneration("failed to load font - %v", err)
	}
	err = pdf.SetFont("roboto", "", 12)
	if err != nil {
		return nil, appError.NewErrPdfGeneration("failed to apply font - %v", err)
	}
	// Add Document header and footer
	pdf.AddHeader(func() {
		pdf.SetXY(xOffset, 40)
		pdf.Cell(nil, "Einsatzmeldung von FireOps")
	})
	pdf.AddFooter(func() {
		pdf.SetXY(xOffset, 800)
		pdf.Cell(nil, time.Now().Format(time.RFC1123))
	})
	// Create page for each event
	for _, event := range events {
		pdf.AddPage()
		pdf.SetXY(xOffset, 100)
		addFieldToPdf(&pdf, "Einsatznummer", event.Num1)
		addFieldToPdf(&pdf, "Kategorie", event.Category)
		addFieldToPdf(&pdf, "Art", event.SubEng)
		addFieldToPdf(&pdf, "Art", event.TypEng)
		addFieldToPdf(&pdf, "Alarmstufe", event.AlarmLev)
		addFieldToPdf(&pdf, "Anrufer", event.CallerName)
		addFieldToPdf(&pdf, "Telefonnummer", event.CallerNumber)
		addFieldToPdf(&pdf, "Ort", event.Location)
		addFieldToPdf(&pdf, "Ortsinfo", event.LocationInfo)
		addFieldToPdf(&pdf, "Info", event.EventAlarmtext)
	}
	return &pdf, nil
}

func addFieldToPdf[T any](pdf *gopdf.GoPdf, name string, value *T) {
	if value != nil {
		pdf.Text(fmt.Sprintf("%s: %v", name, *value))
		pdf.SetXY(xOffset, pdf.GetY()+yOffset)
	}
}
