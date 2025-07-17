package reader

import (
	"bytes"
	"io"
	"github.com/ledongthuc/pdf"
)

func ExtractTextFromPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(&buf, b)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
