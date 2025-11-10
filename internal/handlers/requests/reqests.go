package requests

type SendSmsReq struct {
	To       string `json:"to" validate:"required"`
	Text     string `json:"text" validate:"required"`
	Ttl      int    `json:"ttl" validate:"required"`
	Provider string `json:"provider,omitempty"`
}

type CreateUserReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type CreateServiceReq struct {
	Type          string `json:"type"`
	InitialCredit int64  `json:"initial_credit"`
}

type ChargeReq struct {
	CreditAmount int64 `json:"credit_amount"`
}
