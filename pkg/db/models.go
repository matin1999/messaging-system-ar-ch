package db

import (
	"gorm.io/gorm"
)

type ServiceType string

const (
	ServiceTypeExpress ServiceType = "express"
	ServiceTypeAsync   ServiceType = "async"
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
	Name     string    `gorm:"type:varchar(128);not null;default:''"`
	Password string    `gorm:"type:varchar(128);uniqueIndex;not null;default:''"`
	Services []Service `gorm:"foreignKey:UserID"`
}

type Service struct {
	gorm.Model
	UserID  uint        `gorm:"index;not null"`
	Type    ServiceType `gorm:"type:enum('express','indirect');not null"`
	Status  string      `gorm:"type:varchar(16);not null;default:'active'"`
	Credits uint64      `gorm:"not null;default:0"`
	User    User        `gorm:"references:ID"`
	Sms     []Sms       `gorm:"foreignKey:ServiceId"`
}

type Sms struct {
	gorm.Model
	Content                  string  `gorm:"type:string;not null;"`
	Receptor                 string  `gorm:"type:string;not null;"`
	Status                   string  `gorm:"type:string;not null;"`
	SentTime                 int64   `gorm:"type:int;not null;"`
	Cost                     uint    `gorm:"type:int;default:0;index:idx_service_status_cost;"`
	ServiceProviderName      string  `gorm:"type:string;not null;"`
	ServiceProviderMessageId int     `gorm:"type:int;not null;"`
	ServiceId                uint    `gorm:"references:ID"`
	Service                  Service `gorm:"references:ID"`
}
