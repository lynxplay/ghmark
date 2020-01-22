package pgk

import (
	"fmt"
	"golang.org/x/net/context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const Timeout = time.Minute

func NewChromiumWrapper(port int) *ChromiumWrapper {
	outputDir := ""
	if s, ok := os.LookupEnv("GHMARK_OUTPUT_DIR"); ok {
		outputDir = s
	}

	return &ChromiumWrapper{
		Port:            port,
		OutputDirectory: outputDir,
	}
}

type ChromiumWrapper struct {
	Port            int
	OutputDirectory string
}

func (c *ChromiumWrapper) DownloadPDF(documentId int, fileName string, fallbackOutput string) error {
	if len(c.OutputDirectory) > 0 {
		fallbackOutput = c.OutputDirectory
	}

	cmdContext, _ := context.WithDeadline(context.TODO(), time.Now().Add(Timeout))
	cmd := exec.CommandContext(
		cmdContext,
		"chromium-browser",
		"--no-sandbox",
		"--disable-gpu",
		"--virtual-time-budget=2000",
		"--timeout=6000",
		"--headless",
		fmt.Sprintf("--print-to-pdf=%s", path.Join(fallbackOutput, strings.TrimSuffix(fileName, filepath.Ext(fileName))+".pdf")),
		fmt.Sprintf("http://0.0.0.0:%d/%d", c.Port, documentId),
	)
	return cmd.Run()
}
