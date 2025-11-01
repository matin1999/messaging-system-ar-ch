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
	CreateUserService(userID uint, ServiceType ServiceType, intialCredit int) error
	UpdateServiceCredit(userId uint, serviceId uint, creditAmount int) error
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
	err := d.DBConn.Model(&Service{}).Where("user_id = ?", userID).Find(&svcs).Error
	return svcs, err
}

func (d *DataBaseWrapper) CreateUserService(userID uint, serviceType ServiceType, intialCredit int) error {
	s := &Service{
		UserID:  userID,
		Type:    serviceType,
		Status:  "active",
		Credits: uint64(intialCredit),
	}
	return d.DBConn.Create(s).Error
}

func (d *DataBaseWrapper) CreateUser(name string, password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &User{
		Name:     name,
		Password: string(bytes),
	}
	return d.DBConn.Create(u).Error

}

func (d *DataBaseWrapper) UpdateServiceCredit(userId uint, serviceId uint, creditAmount int) error {
	result := d.DBConn.Model(&Service{}).Where("id = ?", serviceId).Update("credits", creditAmount)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
