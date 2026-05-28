package handlers

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func setupSchedulerTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "scheduler_test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Column{}, &models.Post{}, &models.Setting{}, &models.CustomTag{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取测试数据库连接失败: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	database.DB = db
	config.Cfg = &config.Config{Static: config.StaticConfig{Enable: false, Dir: "public"}}
	InvalidateCache()
}

func TestRunScheduledPublishPublishesOnlyDuePosts(t *testing.T) {
	setupSchedulerTestDB(t)

	now := time.Now()
	past := now.Add(-time.Minute)
	future := now.Add(time.Minute)

	duePost := models.Post{Title: "到期文章", Slug: "due-post", Published: false, ScheduledAt: &past}
	futurePost := models.Post{Title: "未到期文章", Slug: "future-post", Published: false, ScheduledAt: &future}
	draftPost := models.Post{Title: "普通草稿", Slug: "draft-post", Published: false}

	if err := database.DB.Create(&duePost).Error; err != nil {
		t.Fatalf("创建到期文章失败: %v", err)
	}
	if err := database.DB.Create(&futurePost).Error; err != nil {
		t.Fatalf("创建未到期文章失败: %v", err)
	}
	if err := database.DB.Create(&draftPost).Error; err != nil {
		t.Fatalf("创建普通草稿失败: %v", err)
	}

	runScheduledPublish()

	var gotDue models.Post
	if err := database.DB.First(&gotDue, duePost.ID).Error; err != nil {
		t.Fatalf("读取到期文章失败: %v", err)
	}
	if !gotDue.Published {
		t.Fatal("到期文章未被发布")
	}
	if gotDue.ScheduledAt != nil {
		t.Fatal("到期文章发布后 scheduled_at 未清空")
	}

	var gotFuture models.Post
	if err := database.DB.First(&gotFuture, futurePost.ID).Error; err != nil {
		t.Fatalf("读取未到期文章失败: %v", err)
	}
	if gotFuture.Published {
		t.Fatal("未到期文章被提前发布")
	}
	if gotFuture.ScheduledAt == nil {
		t.Fatal("未到期文章 scheduled_at 被错误清空")
	}

	var gotDraft models.Post
	if err := database.DB.First(&gotDraft, draftPost.ID).Error; err != nil {
		t.Fatalf("读取普通草稿失败: %v", err)
	}
	if gotDraft.Published {
		t.Fatal("普通草稿被错误发布")
	}
}
