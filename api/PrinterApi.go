package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/fireops-software/fireops-edge-printer/domain"
	appError "github.com/fireops-software/fireops-edge-printer/error"
	"github.com/uoul/go-common/async"
	"github.com/uoul/go-common/collections"
	"github.com/uoul/go-common/log"
)

type PrinterApi struct {
	ctx          context.Context
	logger       log.ILogger
	printerName  string
	connStr      string
	driver       string
	templateFile string
}

type templateData struct {
	domain.Event
	MapImgBase64 string
}

// IsOnline implements IPrinterApi.
func (p *PrinterApi) IsOnline(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "/usr/bin/lpstat", "-p", p.printerName)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	state := string(out)
	return strings.Contains(state, "idle") || strings.Contains(state, "printing")

}

// PrintPdf implements IPrinterApi.
func (p *PrinterApi) PrintEvents(ctx context.Context, mapSrcLocation string, events []domain.Event, copies int) error {
	// Try setup printer
	if err := p.setup(ctx); err != nil {
		return appError.NewErrPrinter("failed to setup printer - %v", err)
	}
	mapImgs := make([]<-chan async.ActionResult[[]byte], len(events))
	// Create data for template
	templateData := collections.MapSlice(events, func(e domain.Event) templateData {
		return templateData{
			Event: e,
		}
	})
	// Create Map images
	mapCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	for i := 0; i < len(templateData); i++ {
		dest := ""
		if templateData[i].Latitude != nil && templateData[i].Longitude != nil {
			dest = fmt.Sprintf("%v, %v", *templateData[i].Latitude, *templateData[i].Longitude)
		} else if templateData[i].Location != nil {
			dest = *templateData[i].Location
		}
		mapImgs[i] = captureGoogleMapsScreenshot(mapCtx, mapSrcLocation, dest)
	}
	// Create temp file for printing
	mdFile, err := os.CreateTemp("", "job_*.md")
	if err != nil {
		return appError.NewErrPrinter("failed to create md file on filesystem for printing - %v", err)
	}
	// Cleanup
	defer func() {
		mdFile.Close()
		os.Remove(mdFile.Name())
	}()
	// Load Template
	tmpl, err := template.ParseFiles(p.templateFile)
	if err != nil {
		return appError.NewErrPrinter("failed to load template from filesystem - %v", err)
	}
	// Wait for map images
	for i := 0; i < len(templateData); i++ {
		mapResult := <-mapImgs[i]
		if mapResult.Error != nil {
			p.logger.Errorf("Failed to create map for event %v - %v", templateData[i].Num1, mapResult.Error)
			continue // If map creation failed we will go on to print at least the text
		}
		templateData[i].MapImgBase64 = base64.RawStdEncoding.EncodeToString(mapResult.Result)
	}
	// Render Mardown template to tempfile
	if err = tmpl.Execute(mdFile, templateData); err != nil {
		return appError.NewErrPrinter("failed to execute template - %v", err)
	}
	// Close file after pdf content has been written
	mdFile.Close()
	// Convert markdown to html
	cmd1 := exec.CommandContext(
		ctx,
		"/usr/bin/pandoc",
		mdFile.Name(),
		"-o",
		"temp.html",
	)
	if stdOut, err := cmd1.Output(); err != nil {
		return appError.NewErrPrinter("%s - %v", string(stdOut), err)
	} else {
		p.logger.Debugf("Converted mardown to html: %s", stdOut)
	}
	// Create pdf of html
	cmd2 := exec.CommandContext(
		ctx,
		"/usr/bin/chromium",
		"--headless",
		"--disable-gpu",
		"--print-to-pdf=temp.pdf",
		"--no-sandbox",
		"--no-pdf-header-footer",
		"--print-to-pdf-no-header",
		"temp.html",
	)
	if stdOut, err := cmd2.Output(); err != nil {
		return appError.NewErrPrinter("%s - %v", string(stdOut), err)
	} else {
		p.logger.Debugf("Converted html to pdf: %s", stdOut)
	}
	// Print pdf document
	cmd3 := exec.CommandContext(
		ctx,
		"/usr/bin/lpr",
		"-o",
		"portrait",
		"-o",
		"fit-to-page",
		"-o",
		"media=A4",
		"-P",
		p.printerName,
		fmt.Sprintf("-#%d", copies),
		"temp.pdf",
	)
	// Execute print command
	if stdOut, err := cmd3.Output(); err != nil {
		return appError.NewErrPrinter("%s - %v", string(stdOut), err)
	}
	return nil
}

func (p *PrinterApi) setup(ctx context.Context) error {
	// Check if printer already exists
	if printerExists(ctx, p.printerName) {
		return nil
	}
	// Setup printer
	err := exec.CommandContext(
		ctx,
		"/usr/sbin/lpadmin",
		"-p",
		p.printerName,
		"-v",
		p.connStr,
		"-m",
		p.driver,
		"-o",
		"printer-is-shared=false",
		"-E",
	).Run()
	if err == nil {
		p.logger.Infof("Printer %s with connection %s and driver %s has been setup successfully", p.printerName, p.connStr, p.driver)
	}
	return err
}

func captureGoogleMapsScreenshot(ctx context.Context, srcAddr, destAddr string) <-chan async.ActionResult[[]byte] {
	result := make(chan async.ActionResult[[]byte], 1)
	go func() {
		// Create the Google Maps URL
		src := strings.ReplaceAll(srcAddr, " ", "+")
		dest := strings.ReplaceAll(destAddr, " ", "+")
		mapsUrl := fmt.Sprintf("https://maps.google.com/maps?ie=UTF8&output=embed&saddr=%s&daddr=%s&dirflg=d", src, dest)
		// Create HTML with iframe
		html := fmt.Sprintf(`<iframe style="border: 0; width:800px; height:500px; overflow: hidden;" src="%s"></iframe>`, mapsUrl)
		// Create temporary HTML file
		tmpFile, err := os.CreateTemp("", "maps-*.html")
		if err != nil {
			result <- async.NewErrorActionResult[[]byte](
				appError.NewErrIo("failed to create temp file: %v", err),
			)
			return
		}
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.Write([]byte(html)); err != nil {
			result <- async.NewErrorActionResult[[]byte](
				appError.NewErrIo("failed to write data to temp file: %v", err),
			)
			return
		}
		tmpFile.Close()
		// Navigate to file
		fileURL := "file://" + tmpFile.Name()
		// Create temporary file for screenshot
		screenShotFile := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())

		// Build chromium command
		cmd := exec.CommandContext(ctx,
			"/usr/bin/chromium",
			"--headless",
			"--disable-gpu",
			"--no-sandbox",
			"--hide-scrollbars",
			fmt.Sprintf("--screenshot=%s", screenShotFile),
			"--virtual-time-budget=7000", // 7 seconds
			fileURL,
		)
		if stdOut, err := cmd.Output(); err != nil {
			result <- async.NewErrorActionResult[[]byte](appError.NewErrIo("%v - %s", err, string(stdOut)))
			return
		}
		defer os.Remove(screenShotFile)

		// Wait a moment for file to be fully written
		time.Sleep(100 * time.Millisecond)

		// Read the PNG file as []byte
		buf, err := os.ReadFile(screenShotFile)
		if err != nil {
			result <- async.NewErrorActionResult[[]byte](appError.NewErrIo("failed to read image - %v", err))
			return
		}

		result <- async.ActionResult[[]byte]{
			Result: buf,
		}
	}()
	return result
}

func printerExists(ctx context.Context, name string) bool {
	if err := exec.CommandContext(ctx, "/usr/bin/lpstat", "-p", name).Run(); err != nil {
		return false
	}
	return true
}

func WithPrinterDriver(driver string) func(*PrinterApi) {
	return func(pa *PrinterApi) {
		pa.driver = driver
	}
}

func WithPrinterApiTemplate(templateFile string) func(*PrinterApi) {
	return func(pa *PrinterApi) {
		pa.templateFile = templateFile
	}
}

func NewPrinterApi(ctx context.Context, logger log.ILogger, printerName string, connStr string, opts ...func(*PrinterApi)) IPrinterApi {
	p := &PrinterApi{
		ctx:          ctx,
		logger:       logger,
		printerName:  printerName,
		connStr:      connStr,
		driver:       "everywhere",
		templateFile: "templates/events.tpl",
	}
	for _, o := range opts {
		o(p)
	}
	// Try setup printer
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := p.setup(c); err != nil {
		logger.Errorf("failed to setup printer - %v", err)
	}
	// Return printer
	return p
}
