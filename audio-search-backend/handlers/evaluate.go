package handlers

import (
	"audio-search-backend/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func EvaluateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Đường dẫn tới thư mục 80 file test
		testDir := "../audio-feature-extractor/Data_Test"
		instruments, err := os.ReadDir(testDir)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Không tìm thấy thư mục Test"})
		}
		totalQueries := 0
		totalPrecision := 0.0
		totalRecall := 0.0

		for _, instDir := range instruments {
			if !instDir.IsDir() {
				continue
			}
			trueLabel := instDir.Name() // ví dụ: "violin", "cello"
			instPath := filepath.Join(testDir, trueLabel)
			files, _ := os.ReadDir(instPath)
			for _, f := range files {
				if filepath.Ext(f.Name()) != ".wav" {
					continue
				}
				filePath := filepath.Join(instPath, f.Name())
				totalQueries++
				fmt.Printf("⏳ Đang xử lý file [%d/80]: %s\n", totalQueries, filePath)
				// 1. Gọi Python trích xuất cho file Test
				cmd := exec.Command("python", "extract_single.py", filePath)
				output, err := cmd.Output()
				if err != nil {
					continue
				}
				vectorStr := strings.TrimSpace(string(output))
				parts := strings.Split(vectorStr, " ")
				if len(parts) != 19 {
					continue
				}
				pgVectorStr := "[" + strings.Join(parts, ",") + "]"
				// 2. Query Top 5 trong Database
				var results []SearchResult
				query := `
					SELECT instrument
					FROM audio_features
					ORDER BY feature_vector <=> ? ASC
					LIMIT 5
				`
				db.Raw(query, pgVectorStr).Scan(&results)
				// 3. Tính toán
				relevantCount := 0
				for _, r := range results {
					// Nếu kết quả trả về đúng với nhãn thực tế
					if r.Instrument == trueLabel {
						relevantCount++
					}
				}
				// --- ĐẾM CHÍNH XÁC SỐ LƯỢNG FILE CÓ TRONG DB ĐỂ TÍNH RECALL ---
				var dbCount int64
				db.Model(&models.AudioFeature{}).Where("instrument = ?", trueLabel).Count(&dbCount)
				totalRelevantInDB := float64(dbCount)

				// Precision (Độ chính xác) = Số kết quả đúng trong Top 5 / 5
				precision := float64(relevantCount) / 5.0

				// Recall (Độ phủ) = Số kết quả đúng trong Top 5 / Tổng số file đang có thật trong DB
				recall := 0.0
				if totalRelevantInDB > 0 {
					recall = float64(relevantCount) / totalRelevantInDB
				}

				totalPrecision += precision
				totalRecall += recall
			}
		}
		// Tính trung bình (Average Precision / Average Recall)
		var avgPrecision, avgRecall float64
		if totalQueries > 0 {
			avgPrecision = totalPrecision / float64(totalQueries)
			avgRecall = totalRecall / float64(totalQueries)
		}

		return c.JSON(fiber.Map{
			"message":           "Đánh giá hoàn tất",
			"total_test_files":  totalQueries,
			"average_precision": fmt.Sprintf("%.2f%%", avgPrecision*100),
			"average_recall":    fmt.Sprintf("%.2f%%", avgRecall*100),
		})

	}
}
