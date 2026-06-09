package database

import (
	"log"

	"shuaitesteel.com/cms/models"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

const defaultAdminPassword = "G0u8NmtXSsFmDwxDCl"

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
	seedCustomTags()
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
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
	log.Printf("默认管理员已创建: admin / %s", defaultAdminPassword)
}

func seedCustomTags() {
	contactTags := map[string]string{
		"company_name":       "公司名称",
		"company_short_name": "公司简称",
		"company_contact":    "联系人",
		"company_phone":      "联系电话",
		"company_email":      "联系邮箱",
		"company_address":    "联系地址",
		"company_website":    "官网地址",
	}
	for key, name := range contactTags {
		var count int64
		DB.Model(&models.CustomTag{}).Where("key = ?", key).Count(&count)
		if count > 0 {
			continue
		}

		value := ""
		var setting models.Setting
		if err := DB.Where("key = ?", key).First(&setting).Error; err == nil {
			value = setting.Value
		}
		DB.Create(&models.CustomTag{Name: name, Key: key, Value: value})
	}
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
