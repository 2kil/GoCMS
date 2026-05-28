package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func ShowSettings(c *gin.Context) {
	settingMap := loadSettingMap()

	var user models.User
	database.DB.First(&user, c.GetUint("user_id"))

	c.HTML(http.StatusOK, "settings.html", gin.H{
		"settings":       settingMap,
		"frontTemplates": availableFrontTemplates(),
		"nickname":       c.MustGet("nickname"),
		"username":       user.Username,
	})
}

func SaveSettings(c *gin.Context) {
	keys := []string{"site_title", "site_keywords", "site_description", "site_favicon", "site_footer", "front_template", "company_name", "company_short_name", "company_contact", "company_phone", "company_email", "company_address", "company_website"}
	saveSettingsKeys(c, keys)

	InvalidateCache()
	GenerateStatic()
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
	username := c.PostForm("username")
	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")

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

	wantChangeUsername := username != "" && username != user.Username
	wantChangePassword := newPassword != ""

	if !wantChangeUsername && !wantChangePassword {
		c.Redirect(http.StatusFound, "/adm1n/settings")
		return
	}

	if oldPassword == "" {
		showError("修改用户名或密码时必须填写当前密码")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		showError("当前密码错误")
		return
	}

	if wantChangeUsername {
		var existing models.User
		if database.DB.Where("username = ?", username).First(&existing).Error == nil {
			showError("用户名已存在")
			return
		}
		user.Username = username
	}

	if wantChangePassword {
		hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			showError("密码加密失败")
			return
		}
		user.Password = string(hashed)
	}

	database.DB.Save(&user)

	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("nickname", user.Nickname)
	session.Save()

	c.Redirect(http.StatusFound, "/adm1n/settings")
}
