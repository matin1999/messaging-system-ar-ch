package requests

type SendSmsReq struct {
	To   string `json:"to" validate:"required"`
	Text string `json:"text" validate:"required"`
	Provider string `json:"provider,omitempty"`
}


type CreateUserReq struct {
	Name   string `json:"name"`
	Password string `json:"password"`
}

type CreateServiceReq struct {
	Type         string `json:"type"`           
	InitialCharge int64  `json:"initial_charge"` }

type ChargeReq struct {
	ChargeAmount int64 `json:"charge_amount"` 
}