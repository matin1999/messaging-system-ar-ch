package sms

type Service struct {
    provider SmsProvider
}

func NewService(provider SmsProvider) *Service {
    return &Service{provider: provider}
}

func (s *Service) Send(to string, message string) (int,int,error) {
    return s.provider.SendSMS(to, message)
}
