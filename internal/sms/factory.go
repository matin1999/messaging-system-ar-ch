package sms

import (
	"fmt"

	"postchi/internal/sms/providers/kavenegar"
	"postchi/pkg/env"
)

func NewProvider(env *env.Envs,name string) (SmsProvider, error) {
    switch name {
    case "twilio":
        return &kavenegar.SmsProvider{
            ApiKey: env.KAVENEGAR_SMS_API_KEY,
            FromNumber: env.KAVENEGAR_SMS_NUMBER,
        }, nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}
