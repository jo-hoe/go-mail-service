package app

// MailAttributes contains E-Mail attributes
type MailAttributes struct {
	To      []string `json:"to" validate:"min=1"`
	Subject string `json:"subject" validate:"required"`
	Content string `json:"content" validate:"required"`
}
