package handlers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "请输入用户名和密码"})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "用户名或密码错误"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Set("nickname", user.Nickname)
	session.Set("last_login_ip", user.LastLoginIP)
	if user.LastLoginAt != nil {
		session.Set("last_login_at", user.LastLoginAt.Format("2006-01-02 15:04:05"))
	}
	session.Save()

	now := time.Now()
	user.LastLoginIP = c.ClientIP()
	user.LastLoginAt = &now
	database.DB.Save(&user)

	c.Redirect(http.StatusFound, "/adm1n/dashboard")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/adm1n/login")
}
