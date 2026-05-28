package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Log      LogConfig
	Static   StaticConfig
}

type StaticConfig struct {
	Enable bool
	Dir    string
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Path string
}

type SessionConfig struct {
	Secret string
}

type LogConfig struct {
	File string
}

var Cfg *Config

func Init() {
	Cfg = &Config{
		Server:   ServerConfig{Port: "8080"},
		Database: DatabaseConfig{Path: "cms.db"},
		Session:  SessionConfig{Secret: "cms-secret-key-2026"},
		Log:      LogConfig{File: "cms.log"},
		Static:   StaticConfig{Enable: false, Dir: "public"},
	}

	cfgPath := "config.ini"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		exe, _ := os.Executable()
		cfgPath = filepath.Join(filepath.Dir(exe), "config.ini")
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			log.Println("config.ini 不存在，使用默认配置")
			return
		}
	}

	iniFile, err := ini.Load(cfgPath)
	if err != nil {
		log.Printf("加载配置文件失败: %v，使用默认配置", err)
		return
	}

	if sec, err := iniFile.GetSection("server"); err == nil {
		if v := sec.Key("port").String(); v != "" {
			Cfg.Server.Port = v
		}
	}

	if sec, err := iniFile.GetSection("database"); err == nil {
		if v := sec.Key("path").String(); v != "" {
			Cfg.Database.Path = v
		}
	}

	if sec, err := iniFile.GetSection("session"); err == nil {
		if v := sec.Key("secret").String(); v != "" {
			Cfg.Session.Secret = v
		}
	}

	if sec, err := iniFile.GetSection("log"); err == nil {
		if v := sec.Key("file").String(); v != "" {
			Cfg.Log.File = v
		}
	}

	if sec, err := iniFile.GetSection("static"); err == nil {
		if v := sec.Key("enable").String(); v == "true" {
			Cfg.Static.Enable = true
		}
		if v := sec.Key("dir").String(); v != "" {
			Cfg.Static.Dir = v
		}
	}
}

func TemplateDir() string {
	if _, err := os.Stat("templates"); err == nil {
		return "templates"
	}

	if exe, err := os.Executable(); err == nil {
		templatesPath := filepath.Join(filepath.Dir(exe), "templates")
		if _, err := os.Stat(templatesPath); err == nil {
			return templatesPath
		}
	}

	return "templates"

}

func AdminTemplateGlob() string {
	return filepath.Join(TemplateDir(), "admin", "*.html")
}

func FrontTemplateGlob(name string) string {
	return filepath.Join(TemplateDir(), name, "*.html")
}
