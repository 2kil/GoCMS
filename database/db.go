package database

import (
	"log"

	"shuaitesteel.com/cms/models"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := DB.AutoMigrate(&models.User{}, &models.Column{}, &models.Post{}, &models.Setting{}, &models.CustomTag{}); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}

	seedAdmin()
	seedSettings()
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("加密密码失败: %v", err)
	}

	admin := models.User{
		Username: "admin",
		Password: string(hashed),
		Nickname: "管理员",
		IsAdmin:  true,
	}
	DB.Create(&admin)
	log.Println("默认管理员已创建: admin / admin123")
}

func seedSettings() {
	defaults := map[string]string{
		"site_title":         "CMS 内容管理系统",
		"site_keywords":      "CMS,内容管理,网站建设",
		"site_description":   "一个基于 Gin + SQLite 的内容管理系统",
		"site_favicon":       "",
		"site_footer":        "© 2026 CMS. All rights reserved.",
		"front_template":     "index",
		"company_name":       "",
		"company_short_name": "",
		"company_contact":    "",
		"company_phone":      "",
		"company_email":      "",
		"company_address":    "",
		"company_website":    "",
	}
	for key, value := range defaults {
		var count int64
		DB.Model(&models.Setting{}).Where("key = ?", key).Count(&count)
		if count == 0 {
			DB.Create(&models.Setting{Key: key, Value: value})
		}
	}
}
