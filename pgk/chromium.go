package pgk

import (
	"fmt"
	"golang.org/x/net/context"
	"html/template"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const Timeout = time.Second * 5

func NewChromiumWrapper(port int) *ChromiumWrapper {
	outputDir := ""
	if s, ok := os.LookupEnv("GHMARK_OUTPUT_DIR"); ok {
		outputDir = s
	}

	templateString := "{{.FileName}}.pdf"
	if s, ok := os.LookupEnv("GHMARK_OUTPUT_TEMPLATE"); ok {
		templateString = s
	}

	t, _ := template.New("filename").Parse(templateString)

	return &ChromiumWrapper{
		Port:             port,
		OutputDirectory:  outputDir,
		FileNameTemplate: t,
	}
}

type ChromiumWrapper struct {
	Port             int
	OutputDirectory  string
	FileNameTemplate *template.Template
}

func (c *ChromiumWrapper) DownloadPDF(documentId string, fileName string, fallbackOutput string) (string, error) {
	outputDir := fallbackOutput
	if len(c.OutputDirectory) > 0 {
		outputDir = c.OutputDirectory
	}

	builder := strings.Builder{}
	_ = c.FileNameTemplate.Execute(&builder, struct {
		FileName      string
		FileDirectory string
	}{
		FileName:      strings.TrimSuffix(fileName, filepath.Ext(fileName)),
		FileDirectory: fallbackOutput,
	})

	outputPath := path.Join(outputDir, builder.String())

	cmdContext, _ := context.WithDeadline(context.TODO(), time.Now().Add(Timeout))
	cmd := exec.CommandContext(
		cmdContext,
		"chromium-browser",
		"--no-sandbox",
		"--virtual-time-budget=2000",
		"--disable-gpu",
		"--timeout=6000",
		"--headless",
		fmt.Sprintf("--print-to-pdf=%s", outputPath),
		fmt.Sprintf("http://0.0.0.0:%d/%s", c.Port, documentId),
	)
	return outputPath, cmd.Run()
}
