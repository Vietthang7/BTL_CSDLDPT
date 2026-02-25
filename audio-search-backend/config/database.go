package config

import (
	"audio-search-backend/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=123456 dbname=audio_search port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Không thể kết nối PostgreSQL:", err)
	}

	// Tạo extension pgvector
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")

	// Auto migrate bảng
	err = db.AutoMigrate(&models.AudioFeature{})
	if err != nil {
		return nil
	}

	log.Println("✅ Đã kết nối PostgreSQL + pgvector")
	return db
}
