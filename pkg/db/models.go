package db

import (
	"gorm.io/gorm"
)

type ServiceType string

const (
	ServiceTypeExpress  ServiceType = "express"
	ServiceTypeIndirect ServiceType = "indirect"
)

type SmsStatus string

const (
	SmsStatusCreated            SmsStatus = "created"
	SmsStatusQueued             SmsStatus = "queued"
	SmsStatusSent               SmsStatus = "sent"
	SmsStatusDelivered          SmsStatus = "delivered"
	SmsStatusFailed             SmsStatus = "failed"
	SmsStatusInsufficientCharge SmsStatus = "insufficient_charge"
)

type User struct {
	gorm.Model
	UserName string    `gorm:"type:string;size:255;not null;"`
	Password string    `gorm:"type:string;size:255;not null;"`
	Services []Service `gorm:"foreignKey:UserId"`
}

type Service struct {
	gorm.Model
	ServiceSenderNumber string `gorm:"type:string;size:255;not null;"`
	Servicetype         string `gorm:"type:string;size:255;not null;"`
	ChargeAmount        uint64 `gorm:"type:int;default:0"`
	UsedChargeAmount    uint64 `gorm:"type:int;default:0"`
	User                User   `gorm:"references:ID"`
	Sms                 []Sms  `gorm:"foreignKey:ChannelDetectionId"`
}

type Sms struct {
	gorm.Model
	Content                  string  `gorm:"type:string;not null;"`
	SmsStatus                string  `gorm:"type:string;size:255;not null;"`
	Receptor                 string  `gorm:"type:string;not null;"`
	Status                   string  `gorm:"type:string;not null;"`
	SentTime                 int64   `gorm:"type:int;not null;"`
	Cost                     uint    `gorm:"type:int;default:0;index:idx_service_status_cost;"`
	ServiceProviderName      string  `gorm:"type:string;not null;"`
	ServiceProviderMessageId int     `gorm:"type:int;not null;"`
	Service                  Service `gorm:"references:ID"`
}
