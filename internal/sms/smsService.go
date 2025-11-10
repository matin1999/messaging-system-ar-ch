package sms

import "context"

type Service struct {
	provider SmsProvider
}

func NewService(provider SmsProvider) *Service {
	return &Service{provider: provider}
}

func (s *Service) Send(ctx context.Context, to string, message string) (int, int, error) {
	return s.provider.SendSMS(ctx, to, message)
}
