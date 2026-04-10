package services

import (
	"audio-search-backend/models"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

func LoadCSVToDB(db *gorm.DB, filepath string) {
	// Bước 1: Kiểm tra xem DB đã có dữ liệu chưa.
	// Nếu có rồi thì bỏ qua để tránh mỗi lần chạy server lại insert thêm dữ liệu trùng lặp.
	var count int64
	db.Model(&models.AudioFeature{}).Count(&count)
	if count > 0 {
		log.Println("✅ Dữ liệu đã có sẵn trong Database, bỏ qua bước nạp CSV.")
		return
	}
	// Bước 2: Mở file CSV
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("❌ Không thể mở file CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Đọc dòng Header (tiêu đề các cột) và bỏ qua không lưu vào DB
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("❌ Lỗi đọc header CSV: %v", err)
	}
	// Mảng chứa danh sách các bản ghi để insert hàng loạt (batch insert)
	var records []models.AudioFeature
	// Bước 3: Đọc từng dòng của file CSV

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break // Đã đọc đến cuối file
		}
		if err != nil {
			log.Fatalf("❌ Lỗi đọc dòng CSV: %v", err)
		}
		// 19 cột đầu tiên (index từ 0 -> 18) là mảng đặc trưng số thực.
		vectorArray := make([]float32, 19)
		for i := 0; i < 19; i++ {
			val, _ := strconv.ParseFloat(row[i], 32)
			vectorArray[i] = float32(val)
		}
		// Cột 19 là tên nhạc cụ (instrument), cột 20 là tên file (filename)
		instrument := row[19]
		filename := row[20]
		// Đóng gói thành struct Model của Gorm
		// Dùng pgvector.NewVector() để ép kiểu mảng float32 thành dạng Vector của CSDL
		record := models.AudioFeature{
			Filename:      filename,
			Instrument:    instrument,
			FeatureVector: pgvector.NewVector(vectorArray),
		}
		records = append(records, record)
	}

	if len(records) > 0 {
		if err := db.Create(&records).Error; err != nil {
			log.Fatalf("❌ Lỗi khi insert dữ liệu vào DB: %v", err)
		}
		log.Println("✅ Dữ liệu đã được nạp vào Database thành công.")
	}
}
