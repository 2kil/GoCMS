package handlers

import (
	"log"
	"time"

	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func StartPostScheduler() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		runScheduledPublish()
		for range ticker.C {
			runScheduledPublish()
		}
	}()
}

func runScheduledPublish() {
	now := time.Now()
	result := database.DB.Model(&models.Post{}).
		Where("published = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", false, now).
		Updates(map[string]interface{}{
			"published":    true,
			"scheduled_at": nil,
		})
	if result.Error != nil {
		log.Printf("定时发布检查失败: %v", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		return
	}

	InvalidateCache()
	RequestGenerateStatic()
	log.Printf("定时发布完成: %d 篇文章", result.RowsAffected)
}
