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

// LoadCSVToDB nạp dữ liệu đặc trưng (audio_features_normalized.csv) vào bảng audio_features.
// Nếu DB đã có dữ liệu thì bỏ qua bước nạp, chỉ đảm bảo scaler params và HNSW index tồn tại.
func LoadCSVToDB(db *gorm.DB, dataCSVPath string, scalerCSVPath string) {
	// Bước 1: Kiểm tra xem DB đã có dữ liệu chưa.
	var count int64
	db.Model(&models.AudioFeature{}).Count(&count)
	if count > 0 {
		log.Println("✅ Dữ liệu đã có sẵn trong Database, bỏ qua bước nạp CSV.")
		ensureScalerParams(db, scalerCSVPath)
		ensureHNSWIndex(db) // ⭐ Tạo index nếu chưa có
		return
	}
	// Bước 2: Mở file CSV
	file, err := os.Open(dataCSVPath)
	if err != nil {
		log.Fatalf("❌ Không thể mở file CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("❌ Lỗi đọc header CSV: %v", err)
	}

	var records []models.AudioFeature

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("❌ Lỗi đọc dòng CSV: %v", err)
		}

		vectorArray := make([]float32, 19)
		for i := 0; i < 19; i++ {
			val, _ := strconv.ParseFloat(row[i], 32)
			vectorArray[i] = float32(val)
		}

		instrument := row[19]
		filename := row[20]

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

	ensureScalerParams(db, scalerCSVPath)
	ensureHNSWIndex(db) // Tạo HNSW index sau khi load dữ liệu
}

// ensureScalerParams nạp tham số mean/std (scaler_params.csv) vào bảng scaler_params
// nếu bảng này đang trống, để dùng cho việc chuẩn hóa vector truy vấn ở extract_single.py.
func ensureScalerParams(db *gorm.DB, scalerCSVPath string) {
	if scalerCSVPath == "" {
		return
	}

	var count int64
	db.Model(&models.ScalerParam{}).Count(&count)
	if count > 0 {
		return
	}

	file, err := os.Open(scalerCSVPath)
	if err != nil {
		log.Fatalf("❌ Không thể mở file scaler CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("❌ Lỗi đọc header scaler CSV: %v", err)
	}

	var records []models.ScalerParam
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("❌ Lỗi đọc dòng scaler CSV: %v", err)
		}
		if len(row) < 3 {
			continue
		}

		meanVal, errMean := strconv.ParseFloat(row[1], 64)
		stdVal, errStd := strconv.ParseFloat(row[2], 64)
		if errMean != nil || errStd != nil {
			continue
		}

		records = append(records, models.ScalerParam{
			FeatureName: row[0],
			Mean:        meanVal,
			Std:         stdVal,
		})
	}

	if len(records) > 0 {
		if err := db.Create(&records).Error; err != nil {
			log.Fatalf("❌ Lỗi khi insert scaler params vào DB: %v", err)
		}
		log.Println("✅ Đã nạp scaler params vào Database thành công.")
	}
}

// Tạo HNSW index để tối ưu vector search
func ensureHNSWIndex(db *gorm.DB) {
	// Kiểm tra index đã tồn tại chưa
	var indexExists int
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE tablename = 'audio_features' 
		AND indexname = 'idx_audio_features_vector_cosine'
	`).Scan(&indexExists).Error

	if err != nil {
		log.Printf("⚠️  Lỗi kiểm tra index: %v", err)
		return
	}

	if indexExists > 0 {
		log.Println("✅ HNSW index đã tồn tại, bỏ qua bước tạo index.")
		return
	}

	// Tạo HNSW index
	// - m=16: Số connection tối đa mỗi node (16-48 tùy workload)
	// - ef_construction=200: Tham số xây dựng index (100-200)
	// - vector_cosine_ops: Dùng Cosine distance
	createIndexSQL := `
		CREATE INDEX CONCURRENTLY idx_audio_features_vector_cosine 
		ON audio_features 
		USING hnsw (feature_vector vector_cosine_ops)
		WITH (m=16, ef_construction=200);
	`

	if err := db.Exec(createIndexSQL).Error; err != nil {
		log.Printf("❌ Lỗi tạo HNSW index: %v", err)
		return
	}

	log.Println("✅ HNSW index đã được tạo thành công!")
	log.Println("   - m=16: Số connection mỗi node")
	log.Println("   - ef_construction=200: Construction parameter")
	log.Println("   - vector_cosine_ops: Cosine Distance")
	log.Println("   🚀 Tìm kiếm sẽ nhanh hơn 10x!")
}
