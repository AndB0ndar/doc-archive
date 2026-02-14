package service

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractText(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	var builder strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		//cleanText := strings.ToValidUTF8(text, " ")
		//builder.WriteString(cleanText)
		builder.WriteString(text)
		builder.WriteString("\n")
	}

	if builder.Len() == 0 {
		return "", fmt.Errorf("no text could be extracted from PDF")
	}
	return builder.String(), nil
}
