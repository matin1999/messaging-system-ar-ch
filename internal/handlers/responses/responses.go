package responses

type SendSmsResp struct {
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
}