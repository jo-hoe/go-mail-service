package app

// Service forwards E-Mail to a set of receivers
type MailService interface {
	SendMail(attributes MailAttributes) error
}