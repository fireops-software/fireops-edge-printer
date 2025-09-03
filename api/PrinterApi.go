package api

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/fireops-software/fireops-edge-printer/domain"
	appError "github.com/fireops-software/fireops-edge-printer/error"
	"github.com/uoul/go-common/log"
)

const (
	PRINT_CMD = "/usr/bin/lpr -o portrait -o fit-to-page -o media=A4 -P %s -#%d %s"
)

type PrinterApi struct {
	logger       log.ILogger
	printerName  string
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

func WithPrinterApiTemplate(templateFile string) func(*PrinterApi) {
	return func(pa *PrinterApi) {
		pa.templateFile = templateFile
	}
}

func NewPrinterApi(logger log.ILogger, printerName string, opts ...func(*PrinterApi)) IPrinterApi {
	p := &PrinterApi{
		logger:       logger,
		printerName:  printerName,
		templateFile: "templates/events.tpl",
	}
	for _, o := range opts {
		o(p)
	}
	return p
}
