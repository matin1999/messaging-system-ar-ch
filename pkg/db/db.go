package db

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DataBaseInterface interface {
	DB() *gorm.DB
	CreateUser(name string, password string) error
	GetUserServices(userID uint) ([]Service, error)
	CreateUserService(userID uint, typ ServiceType) (*Service, error)
}

type DataBaseWrapper struct {
	DBConn *gorm.DB
}

func (d *DataBaseWrapper) DB() *gorm.DB { return d.DBConn }

func Init(dsn string) (DataBaseInterface, error) {
	if dsn == "" {
		return nil, errors.New("db: empty DSN")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&User{}, &Service{}, &Sms{}); err != nil {
		return nil, err
	}
	return &DataBaseWrapper{DBConn: db}, nil
}

func (d *DataBaseWrapper) GetUserServices(userID uint) ([]Service, error) {
	var svcs []Service
	err := d.DBConn.Where("user_id = ?", userID).Find(&svcs).Error
	return svcs, err
}

func (d *DataBaseWrapper) CreateUserService(userID uint, typ ServiceType) (*Service, error) {
	s := &Service{
		UserID:  userID,
		Type:    typ,
		Status:  "active",
		Credits: 0,
	}
	return s, d.DBConn.Create(s).Error
}

func (d *DataBaseWrapper) CreateUser(name string, password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

}
