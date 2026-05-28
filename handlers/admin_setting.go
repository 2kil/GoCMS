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

	if username != "" && username != user.Username {
		if oldPassword == "" {
			c.Redirect(http.StatusFound, "/adm1n/settings")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			c.Redirect(http.StatusFound, "/adm1n/settings")
			return
		}
		var existing models.User
		if database.DB.Where("username = ?", username).First(&existing).Error == nil {
			var settings []models.Setting
			database.DB.Find(&settings)
			settingMap := make(map[string]string)
			for _, s := range settings {
				settingMap[s.Key] = s.Value
			}
			c.HTML(http.StatusOK, "settings.html", gin.H{
				"error":          "用户名已存在",
				"settings":       settingMap,
				"frontTemplates": availableFrontTemplates(),
				"username":       user.Username,
				"nickname":       c.MustGet("nickname"),
			})
			return
		}
		user.Username = username
	}

	if newPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			c.Redirect(http.StatusFound, "/adm1n/settings")
			return
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			c.Redirect(http.StatusFound, "/adm1n/settings")
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
