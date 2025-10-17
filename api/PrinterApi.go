package api

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/fireops-software/fireops-edge-printer/domain"
	appError "github.com/fireops-software/fireops-edge-printer/error"
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
func (p *PrinterApi) PrintEvents(ctx context.Context, events []domain.Event, copies int) error {
	// Try setup printer
	if err := p.setup(ctx); err != nil {
		return appError.NewErrPrinter("failed to setup printer - %v", err)
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
	// Render Mardown template to tempfile
	if err = tmpl.Execute(mdFile, events); err != nil {
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
