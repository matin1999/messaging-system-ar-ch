package requests

type SendSmsReq struct {
	To   string `json:"to" validate:"required"`
	Text string `json:"text" validate:"required"`
	Provider string `json:"provider,omitempty"`
}