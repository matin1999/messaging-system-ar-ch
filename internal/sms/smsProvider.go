package sms

import "context"

type SmsProvider interface {
	SendSMS(ctx context.Context, to string, message string) (int, int, error)
	GetName() string
}
