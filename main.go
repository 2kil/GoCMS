package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/handlers"
	"shuaitesteel.com/cms/middleware"
)

var version = "dev"

func main() {
	args := os.Args[1:]
	cmd := "serve"
	if len(args) > 0 {
		cmd = args[0]
	}
	if cmd == "version" || cmd == "-v" || cmd == "--version" {
		fmt.Println(version)
		return
	}
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		printUsage()
		return
	}
	if isCLICommand(cmd) && len(args) > 1 && isHelpArg(args[1]) {
		printCommandHelp(cmd)
		return
	}

	config.Init()
	if cmd == "refresh" {
		requestRefresh()
		return
	}
	setupLog()
	database.Init(config.Cfg.Database.Path)

	switch cmd {
	case "serve":
		runServer()
	case "static", "generate-static":
		handlers.GenerateStatic()
		log.Printf("静态页面生成完成: %s", config.Cfg.Static.Dir)
	case "migrate":
		log.Printf("数据库迁移完成: %s", config.Cfg.Database.Path)
	case "user", "post", "column", "tag", "setting", "upload", "db", "reset-admin":
		runCLI(args)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func isCLICommand(cmd string) bool {
	switch cmd {
	case "user", "post", "column", "tag", "setting", "upload", "db", "reset-admin":
		return true
	default:
		return false
	}
}

func setupLog() {
	if config.Cfg.Log.File == "" {
		gin.DefaultWriter = os.Stdout
		return
	}

	if f, err := os.OpenFile(config.Cfg.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
		log.SetOutput(gin.DefaultWriter)
	} else {
		log.Printf("创建日志文件失败: %v，仅输出到控制台", err)
		gin.DefaultWriter = os.Stdout
		log.SetOutput(os.Stdout)
	}
}

func runServer() {
	handlers.GenerateStatic()
	handlers.StartStaticGenerator()
	handlers.StartPostScheduler()
	startRefreshWatcher()

	r := gin.Default()

	r.SetFuncMap(handlers.TemplateFuncsForGin())

	r.LoadHTMLGlob(config.AdminTemplateGlob())

	r.Static("/static", config.ResolvePath("static"))
	r.Static("/upload", config.ResolvePath("upload"))
	r.Static("/admin-assets/css", filepath.Join(config.TemplateDir(), "admin", "css"))

	store := cookie.NewStore([]byte(config.Cfg.Session.Secret))
	r.Use(sessions.Sessions("cms_session", store))

	admin := r.Group("/adm1n")
	{
		admin.GET("/login", handlers.ShowLogin)
		admin.POST("/login", handlers.Login)
		admin.GET("/logout", handlers.Logout)

		auth := admin.Group("")
		auth.Use(middleware.RequireAuth())
		auth.Use(middleware.RequireCSRF())
		auth.Use(func(c *gin.Context) {
			session := sessions.Default(c)
			c.Set("user_id", session.Get("user_id"))
			c.Set("username", session.Get("username"))
			c.Set("nickname", session.Get("nickname"))
			c.Next()
		})
		{
			auth.GET("/dashboard", handlers.Dashboard)
			auth.POST("/uploads/images", handlers.UploadImage)
			auth.POST("/uploads/covers", handlers.UploadCoverImage)
			auth.POST("/uploads/upload", handlers.UploadImages)
			auth.GET("/uploads", handlers.ListUploads)
			auth.POST("/uploads/rename", handlers.RenameUpload)
			auth.POST("/uploads/delete", handlers.DeleteUpload)
			auth.POST("/uploads/delete-selected", handlers.DeleteUploads)

			auth.GET("/posts", handlers.ListPosts)
			auth.GET("/posts/edit/:id", handlers.ShowPostEdit)
			auth.POST("/posts/save/:id", handlers.SavePost)
			auth.POST("/posts/delete/:id", handlers.DeletePost)
			auth.POST("/posts/toggle/:id", handlers.TogglePost)

			auth.GET("/columns", handlers.ListColumns)
			auth.GET("/columns/edit/:id", handlers.ShowColumnEdit)
			auth.POST("/columns/save/:id", handlers.SaveColumn)
			auth.POST("/columns/reorder", handlers.ReorderColumns)
			auth.POST("/columns/delete/:id", handlers.DeleteColumn)

			auth.GET("/tags", handlers.ListTags)
			auth.GET("/tags/edit/:id", handlers.ShowTagEdit)
			auth.POST("/tags/save/:id", handlers.SaveTag)
			auth.POST("/tags/delete/:id", handlers.DeleteTag)

			auth.GET("/settings", handlers.ShowSettings)
			auth.POST("/settings", handlers.SaveSettings)
			auth.POST("/settings/account", handlers.SaveAccount)
		}
	}

	r.NoRoute(handlers.ServePublic)

	port := config.Cfg.Server.Port

	log.Printf("服务器启动: http://localhost:%s", port)
	log.Printf("管理后台: http://localhost:%s/adm1n/login", port)
	log.Printf("默认账号: admin / G0u8NmtXSsFmDwxDCl")
	r.Run(":" + port)
}

func printUsage() {
	fmt.Println(`GoCMS CLI

用法:
  cms [command]

命令:
  serve                         启动 Web 服务（默认）
  refresh                       请求运行中的服务热更新 public 静态文件
  static                        重新生成前台静态页面
  generate-static               同 static
  migrate                       初始化并迁移数据库
  user <action>                 后台用户维护
  post <action>                 文章维护
  column <action>               栏目维护
  tag <action>                  自定义标签维护
  setting <action>              网站设置维护
  upload <action>               上传图片维护
  db <action>                   数据库初始化、结构输出和修复
  reset-admin <user> <password> 重置已有管理员密码（兼容旧命令）
  version                       输出版本号
  help                          显示帮助

不带 command 时等同于 cms serve。执行 cms <command> help 查看子命令。`)
}
