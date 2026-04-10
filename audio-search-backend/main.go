package main

import (
	"audio-search-backend/config"
	"audio-search-backend/handlers"
	"audio-search-backend/services"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Kết nối Database
	db := config.ConnectDB()

	// 2. Load CSV vào DB (chạy 1 lần, tự bỏ qua nếu đã có dữ liệu)
	services.LoadCSVToDB(db, "../audio-feature-extractor/audio_features.csv")

	// 3. Khởi tạo Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB cho upload file audio
	})
	// THÊM DÒNG NÀY ĐỂ HIỂN THỊ GIAO DIỆN WEB TỪ THƯ MỤC "public"
	app.Static("/", "./public")

	//// 4. Routes
	api := app.Group("/api")
	{
		api.Post("/search", handlers.SearchHandler(db))
		api.Get("/evaluate", handlers.EvaluateHandler(db))
	}

	log.Println("🚀 Server đang chạy tại http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
