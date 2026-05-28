package main

import (
	"io"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/handlers"
	"shuaitesteel.com/cms/middleware"
)

func main() {
	config.Init()
	database.Init(config.Cfg.Database.Path)
	handlers.GenerateStatic()
	handlers.StartPostScheduler()

	if f, err := os.Create(config.Cfg.Log.File); err == nil {
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	} else {
		log.Printf("创建日志文件失败: %v，仅输出到控制台", err)
		gin.DefaultWriter = os.Stdout
	}

	r := gin.Default()

	r.SetFuncMap(handlers.TemplateFuncsForGin())

	r.LoadHTMLGlob(config.AdminTemplateGlob())

	r.Static("/static", "./static")
	r.Static("/admin-assets/css", "./templates/admin/css")

	store := cookie.NewStore([]byte(config.Cfg.Session.Secret))
	r.Use(sessions.Sessions("cms_session", store))

	admin := r.Group("/adm1n")
	{
		admin.GET("/login", handlers.ShowLogin)
		admin.POST("/login", handlers.Login)
		admin.GET("/logout", handlers.Logout)

		auth := admin.Group("")
		auth.Use(middleware.RequireAuth())
		auth.Use(func(c *gin.Context) {
			session := sessions.Default(c)
			c.Set("user_id", session.Get("user_id"))
			c.Set("username", session.Get("username"))
			c.Set("nickname", session.Get("nickname"))
			c.Next()
		})
		{
			auth.GET("/dashboard", handlers.Dashboard)

			auth.GET("/posts", handlers.ListPosts)
			auth.GET("/posts/edit/:id", handlers.ShowPostEdit)
			auth.POST("/posts/save/:id", handlers.SavePost)
			auth.GET("/posts/delete/:id", handlers.DeletePost)
			auth.GET("/posts/toggle/:id", handlers.TogglePost)

			auth.GET("/columns", handlers.ListColumns)
			auth.GET("/columns/edit/:id", handlers.ShowColumnEdit)
			auth.POST("/columns/save/:id", handlers.SaveColumn)
			auth.POST("/columns/reorder", handlers.ReorderColumns)
			auth.GET("/columns/delete/:id", handlers.DeleteColumn)

			auth.GET("/tags", handlers.ListTags)
			auth.GET("/tags/edit/:id", handlers.ShowTagEdit)
			auth.POST("/tags/save/:id", handlers.SaveTag)
			auth.GET("/tags/delete/:id", handlers.DeleteTag)

			auth.GET("/settings", handlers.ShowSettings)
			auth.POST("/settings", handlers.SaveSettings)
			auth.POST("/settings/account", handlers.SaveAccount)
		}
	}

	r.NoRoute(handlers.ServePublic)

	port := config.Cfg.Server.Port

	log.Printf("服务器启动: http://localhost:%s", port)
	log.Printf("管理后台: http://localhost:%s/adm1n/login", port)
	log.Printf("默认账号: admin / admin123")
	r.Run(":" + port)
}
