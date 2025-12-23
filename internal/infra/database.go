package infra

import (
	"fmt"
	"go-ecommerce-app/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg config.AppConfig) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	log.Printf("Connecting to database: host=%s port=%s dbname=%s user=%s", cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("✅ Database connected successfully: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
