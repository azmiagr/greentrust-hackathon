package mail

import (
	"bytes"
	"embed"
	htmlTemplate "html/template"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

type VerificationEmailData struct {
	Name             string
	Code             string
	VerificationLink string
	ExpiryMinutes    int
	Year             int
}

var templates *htmlTemplate.Template

func init() {
	var err error
	templates, err = htmlTemplate.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic("failed to parse email templates: " + err.Error())
	}
}

func RenderVerificationEmail(data VerificationEmailData) (string, error) {
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	if data.ExpiryMinutes == 0 {
		data.ExpiryMinutes = 10
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "verification.html", data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type PreOrderEmailData struct {
	Name         string
	PreOrderName string
	DiscountCode string
	Year         int
}

func RenderPreOrderEmail(data PreOrderEmailData) (string, error) {
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "preorder.html", data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type SelectionRejectedEmailData struct {
	Name       string
	ClassName  string
	ReviewNote string
	Year       int
}

func RenderSelectionRejectedEmail(data SelectionRejectedEmailData) (string, error) {
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "selection_rejected.html", data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type SelectionAcceptedEmailData struct {
	Name      string
	ClassName string
	Year      int
}

func RenderSelectionAcceptedEmail(data SelectionAcceptedEmailData) (string, error) {
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "selection_accepted.html", data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
