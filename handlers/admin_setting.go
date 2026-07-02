package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func ShowSettings(c *gin.Context) {
	settingMap := loadSettingMap()

	c.HTML(http.StatusOK, "settings.html", gin.H{
		"settings":       settingMap,
		"frontTemplates": availableFrontTemplates(),
		"nickname":       c.MustGet("nickname"),
	})
}

func SaveSettings(c *gin.Context) {
	keys := []string{"site_title", "site_keywords", "site_description", "site_favicon", "site_footer", "front_template"}
	saveSettingsKeys(c, keys)

	InvalidateCache()
	RequestGenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/settings")
}

func loadSettingMap() map[string]string {
	var settings []models.Setting
	database.DB.Find(&settings)
	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}
	settingMap["front_template"] = selectedFrontTemplate(settingMap)
	return settingMap
}

func saveSettingsKeys(c *gin.Context, keys []string) {
	for _, key := range keys {
		value := c.PostForm(key)
		var setting models.Setting
		result := database.DB.Where("key = ?", key).First(&setting)
		if result.Error != nil {
			database.DB.Create(&models.Setting{Key: key, Value: value})
		} else {
			setting.Value = value
			database.DB.Save(&setting)
		}
	}
}

func SaveAccount(c *gin.Context) {
	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	userID := c.GetUint("user_id")
	var user models.User
	database.DB.First(&user, userID)

	showError := func(msg string) {
		settingMap := loadSettingMap()
		c.HTML(http.StatusOK, "settings.html", gin.H{
			"error":          msg,
			"settings":       settingMap,
			"frontTemplates": availableFrontTemplates(),
			"username":       user.Username,
			"nickname":       c.MustGet("nickname"),
		})
	}

	wantChangePassword := newPassword != ""

	if !wantChangePassword {
		c.Redirect(http.StatusFound, "/adm1n/settings")
		return
	}

	if oldPassword == "" {
		showError("修改密码时必须填写当前密码")
		return
	}
	if newPassword != confirmPassword {
		showError("两次输入的新密码不一致")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		showError("当前密码错误")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		showError("密码加密失败")
		return
	}
	user.Password = string(hashed)

	database.DB.Save(&user)

	c.Redirect(http.StatusFound, "/adm1n/settings")
}
