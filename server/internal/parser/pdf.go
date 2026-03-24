package parser

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "github.com/AbdullahAlzariqi/pdf"
)

func ExtractFromPDF(pdfPath string) (string, error) {
    text, err := extractPDF(pdfPath)
    if err == nil {
        return text, nil
    }
    return extractWithGoLibrary(pdfPath)
}

func extractPDF(pdfPath string) (string, error) {
    pdftotextPath, err := exec.LookPath("pdftotext")
    if err != nil {
        return "", fmt.Errorf("pdftotext not found in PATH: %w", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

	// Launching: pdftotext
	// -layout – preserves positioning (useful for tables and columns);
	// -nopgbrk – does not insert page breaks;
	// -enc UTF-8 – explicitly specifies the encoding.
	// "-" - indicates to output text to stdout
	cmd := exec.CommandContext(
		ctx, pdftotextPath, "-layout", "-enc", "UTF-8", pdfPath, "-",
	)

    output, err := cmd.Output()
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return "", fmt.Errorf(
				"pdftotext failed (exit %d): %s",
				exitErr.ExitCode(),
				string(exitErr.Stderr),
			)
        }
        return "", fmt.Errorf("failed to run pdftotext: %w", err)
    }

    return strings.TrimSpace(string(output)), nil
}

func extractWithGoLibrary(pdfPath string) (string, error) {
    file, reader, err := pdf.Open(pdfPath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    var builder strings.Builder

    for pageNum := 1; pageNum <= reader.NumPage(); pageNum++ {
        page := reader.Page(pageNum)
        if page.V.IsNull() {
            continue
        }

        rows, err := page.GetTextByRow()
        if err != nil {
            return "", err
        }

        for _, row := range rows {
            var lineBuilder strings.Builder
            for _, word := range row.Content {
                lineBuilder.WriteString(word.S)
                lineBuilder.WriteByte(' ')
            }
            line := strings.TrimRight(lineBuilder.String(), " ")
            builder.WriteString(line)
            builder.WriteByte('\n')
        }
    }

    return builder.String(), nil
}
