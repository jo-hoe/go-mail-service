package mail

// MailAttributes contains E-Mail attributes
type MailAttributes struct {
	To          string `json:"to" validate:"required"`
	Subject     string `json:"subject" validate:"required"`
	HtmlContent string `json:"content" validate:"required"`
	From        string `json:"from,omitempty"`
	FromName    string `json:"fromName,omitempty"`
}
