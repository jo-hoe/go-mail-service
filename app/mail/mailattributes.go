package mail

// MailAttributes contains E-Mail attributes
type MailAttributes struct {
	To          []string `json:"to" validate:"min=1"`
	Subject     string   `json:"subject" validate:"required"`
	HtmlContent string   `json:"content" validate:"required"`
	From        string   `json:"from"`
	FromName    string   `json:"fromName"`
}
