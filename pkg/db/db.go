package db

import (


	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DataBaseInterface interface {
}

type DataBaseWrapper struct {
	PG *gorm.DB
}

func Init(dsn string) (DataBaseInterface, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if db.AutoMigrate(&User{}, &Service{}, &Sms{}) != nil {
		return nil, err
	}
	return &DataBaseWrapper{PG: db}, nil

}

