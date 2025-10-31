package sms

type SmsProvider interface {
	SendSMS(to string, message string) (int, int, error)
	GetName() string
}
