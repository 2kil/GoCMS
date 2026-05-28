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
	var posts []models.Post
	if err := database.DB.Where("published = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", false, now).Find(&posts).Error; err != nil {
		log.Printf("定时发布检查失败: %v", err)
		return
	}
	if len(posts) == 0 {
		return
	}

	for i := range posts {
		posts[i].Published = true
		posts[i].ScheduledAt = nil
		if err := database.DB.Save(&posts[i]).Error; err != nil {
			log.Printf("定时发布文章失败 id=%d: %v", posts[i].ID, err)
		}
	}

	InvalidateCache()
	GenerateStatic()
	log.Printf("定时发布完成: %d 篇文章", len(posts))
}
