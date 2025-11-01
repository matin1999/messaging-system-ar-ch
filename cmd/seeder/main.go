package seeder

import (
	"fmt"
	"os"
	"time"

	"postchi/pkg/db"
	"postchi/pkg/env"
	"postchi/pkg/logger"

	"gorm.io/gorm"
)

func main() {
	envs := env.ReadEnvs()
	log, err := logger.Init(&envs)
	if err != nil {
		log.StdLog("error", fmt.Sprintf("[seeder] logger init failed: %v", err))
		os.Exit(1)
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.StdLog("error", "[seeder] DB_DSN must be set")
		os.Exit(1)
	}

	store, err := db.Init(dsn)
	if err != nil {
		log.StdLog("error", fmt.Sprintf("[seeder] db init failed: %v", err))
		os.Exit(1)
	}
	g := store.DB()

	if err := g.Transaction(func(tx *gorm.DB) error {
		users := []db.User{
			{Name: "Alice", APIKey: "api_alice_123"},
			{Name: "Bob", APIKey: "api_bob_123"},
			{Name: "Carol", APIKey: "api_carol_123"},
		}

		for i := range users {
			if err := tx.Where("api_key = ?", users[i].APIKey).FirstOrCreate(&users[i]).Error; err != nil {
				return fmt.Errorf("create user %s: %w", users[i].APIKey, err)
			}
		}

		for i := range users {
			u := users[i]

			if err := firstOrCreateService(tx, u.ID, db.ServiceTypeExpress, 1_000); err != nil {
				return err
			}
			if err := firstOrCreateService(tx, u.ID, db.ServiceTypeIndirect, 5_000); err != nil {
				return err
			}
		}

		if err := seedSampleQueuedSms(tx, users[0]); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.StdLog("error", fmt.Sprintf("[seeder] transaction failed: %v", err))
		os.Exit(1)
	}

	log.StdLog("info", "[seeder] done")
}

func firstOrCreateService(tx *gorm.DB, userID uint, typ db.ServiceType, minCredits int64) error {
	var svc db.Service
	if err := tx.Where("user_id = ? AND type = ?", userID, typ).First(&svc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			svc = db.Service{
				UserID:  userID,
				Type:    typ,
				Status:  "active",
				Credits: minCredits,
			}
			if err := tx.Create(&svc).Error; err != nil {
				return fmt.Errorf("create service (user=%d,type=%s): %w", userID, typ, err)
			}
			return nil
		}
		return fmt.Errorf("lookup service (user=%d,type=%s): %w", userID, typ, err)
	}
	// Service exists: ensure it has at least minCredits
	if svc.Credits < minCredits {
		if err := tx.Model(&db.Service{}).
			Where("id = ?", svc.ID).
			Update("credits", minCredits).Error; err != nil {
			return fmt.Errorf("bump credits (svc=%d): %w", svc.ID, err)
		}
	}
	return nil
}

// Seeds a couple queued SMS examples for a user's indirect service (if present)
func seedSampleQueuedSms(tx *gorm.DB, user db.User) error {
	// Find indirect service for this user
	var svc db.Service
	if err := tx.Where("user_id = ? AND type = ?", user.ID, db.ServiceTypeIndirect).First(&svc).Error; err != nil {
		// If missing, it's not critical for seeding; just return nil
		return nil
	}

	now := time.Now().UTC().Unix()

	queued := []db.Sms{
		{
			Content:                  "Hello from seeder (1)!",
			SmsStatus:                "queued",
			Receptor:                 "+989120000001",
			Status:                   "queued",
			SentTime:                 now,
			Cost:                     0,
			ServiceProviderName:      "",
			ServiceProviderMessageId: 0,
			Service:                  svc,
		},
		{
			Content:                  "Hello from seeder (2)!",
			SmsStatus:                "queued",
			Receptor:                 "+989120000002",
			Status:                   "queued",
			SentTime:                 now,
			Cost:                     0,
			ServiceProviderName:      "",
			ServiceProviderMessageId: 0,
			Service:                  svc,
		},
	}

	for _, s := range queued {
		var existing db.Sms
		err := tx.Where("receptor = ? AND status = 'queued'", s.Receptor).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&s).Error; err != nil {
				return fmt.Errorf("create queued sms (%s): %w", s.Receptor, err)
			}
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}
