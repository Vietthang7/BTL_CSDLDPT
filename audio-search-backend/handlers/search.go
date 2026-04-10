package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SearchResult là cấu trúc hiển thị danh sách kết quả
type SearchResult struct {
	ID         uint    `json:"id"`
	Filename   string  `json:"filename"`
	Instrument string  `json:"instrument"`
	Distance   float64 `json:"distance"` // Khoảng cách Cosine (càng nhỏ càng giống)
}

func SearchHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Nhận file upload từ người dùng (từ thẻ <input type="file" name="audio">)
		file, err := c.FormFile("audio")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Không tìm thấy file tải lên"})
		}
		// Lưu file tạm thời để Python đọc
		tempPath := fmt.Sprintf("./temp_%s", file.Filename)
		if err := c.SaveFile(file, tempPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Lỗi lưu file tạm"})
		}
		defer os.Remove(tempPath) // Đảm bảo luôn xóa file tạm khi chạy xong
		// 2. Gọi script Python trích xuất Vector 19 chiều
		cmd := exec.Command("python", "extract_single.py", tempPath)
		output, err := cmd.Output()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Lỗi khi trích xuất đặc trưng bằng Python"})
		}
		// Cắt chuỗi output thành mảng
		vectorStr := strings.TrimSpace(string(output))
		parts := strings.Split(vectorStr, " ")
		if len(parts) != 19 {
			return c.Status(500).JSON(fiber.Map{"error": "Trích xuất thất bại, Vector không đủ 19 chiều"})
		}
		// Định dạng vector thành chuỗi dành riêng cho PostgreSQL pgvector: "[v1, v2, ..., v19]"
		pgVectorStr := "[" + strings.Join(parts, ",") + "]"
		// 3. Truy vấn Database tính toán Cosine Distance
		// Toán tử <=> trong pgvector đại diện cho việc tính Cosine Distance
		var results []SearchResult
		query := `
			SELECT id, filename, instrument, feature_vector <=> ? AS distance
			FROM audio_features
			ORDER BY distance ASC
			LIMIT 5
		`
		if err := db.Raw(query, pgVectorStr).Scan(&results).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Lỗi truy vấn Database"})
		}
		// 4. Trả kết quả về cho client
		return c.JSON(fiber.Map{
			"message": "Tìm kiếm thành công",
			"results": results,
		})
	}
}
