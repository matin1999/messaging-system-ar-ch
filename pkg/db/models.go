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
	SmsStatusQueued    SmsStatus = "queued"
	SmsStatusSent      SmsStatus = "sent"
	SmsStatusDelivered SmsStatus = "delivered"
	SmsStatusFailed    SmsStatus = "failed"
)

type User struct {
	gorm.Model
	Name   string `gorm:"type:varchar(128);not null;default:''"`
	APIKey string `gorm:"type:varchar(128);uniqueIndex;not null;default:''"`
}

type Service struct {
	gorm.Model
	UserID  uint        `gorm:"index;not null"`
	Type    ServiceType `gorm:"type:enum('express','indirect');not null"`
	Status  string      `gorm:"type:varchar(16);not null;default:'active'"`
	Credits int64       `gorm:"not null;default:0"`
	User    User        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Sms     []Sms       `gorm:"foreignKey:ChannelDetectionId"`
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
