package kavenegar

import (
	"context"

	"github.com/kavenegar/kavenegar-go"
)

type SmsProvider struct {
	ApiKey     string
	FromNumber string
}

func (p *SmsProvider) SendSMS(ctx context.Context, to string, message string) (int, int, error) {
	api := kavenegar.New(p.ApiKey)
	if res, err := api.Message.Send(p.FromNumber, []string{to}, message, nil); err != nil {
		return 0, 0, err
	} else {
		return int(res[0].Status), res[0].MessageID, nil
	}
}

func (p *SmsProvider) GetName() string {
	return "kavenegar"
}
