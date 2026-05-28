package handlers

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func getSettings() map[string]string {
	return frontCache.getSettings()
}

func loadColumnsFlat() []models.Column {
	return frontCache.getColumns()
}

func staticFile(path string) string {
	if !config.Cfg.Static.Enable {
		return ""
	}
	full := filepath.Join(config.Cfg.Static.Dir, path)
	if _, err := os.Stat(full); err == nil {
		return full
	}
	return ""
}

func serveStaticOr(c *gin.Context, staticPath string, fn func()) {
	if p := staticFile(staticPath); p != "" {
		c.File(p)
		return
	}
	fn()
}

func renderFront(c *gin.Context, status int, data map[string]interface{}) {
	settings := getSettings()
	if _, ok := data["settings"]; !ok {
		data["settings"] = settings
	}
	tpl := frontTemplates(settings)
	if tpl == nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "index.html", data); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func ServePublic(c *gin.Context) {
	requestPath := c.Request.URL.Path
	if strings.HasPrefix(requestPath, "/adm1n") {
		c.Status(http.StatusNotFound)
		return
	}

	cleanPath := path.Clean("/" + requestPath)
	if cleanPath == "/" {
		cleanPath = "/index.html"
	}

	filePath := filepath.Join(config.Cfg.Static.Dir, filepath.FromSlash(strings.TrimPrefix(cleanPath, "/")))
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		c.File(filePath)
		return
	}

	c.Status(http.StatusNotFound)
}

func Home(c *gin.Context) {
	serveStaticOr(c, "index.html", func() {
		var posts []models.Post
		database.DB.Where("published = ?", true).Preload("Column").Order("id desc").Limit(10).Find(&posts)

		renderFront(c, http.StatusOK, gin.H{
			"posts":       posts,
			"columns":     loadColumnsFlat(),
			"custom_tags": frontCache.getCustomTags(),
		})
	})
}

func ShowPost(c *gin.Context) {
	slug := strings.TrimSuffix(c.Param("slug"), ".html")

	serveStaticOr(c, filepath.Join("post", slug+".html"), func() {
		var post models.Post
		result := database.DB.Where("slug = ? AND published = ?", slug, true).Preload("Column").Preload("User").First(&post)
		if result.Error != nil {
			renderFront(c, http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}

		renderFront(c, http.StatusOK, gin.H{
			"post":        post,
			"columns":     loadColumnsFlat(),
			"custom_tags": frontCache.getCustomTags(),
		})
	})
}

func ListPostsByColumn(c *gin.Context) {
	slug := c.Param("slug")

	serveStaticOr(c, filepath.Join("column", slug+".html"), func() {
		var column models.Column
		if err := database.DB.Where("slug = ?", slug).First(&column).Error; err != nil {
			renderFront(c, http.StatusNotFound, gin.H{"error": "栏目不存在"})
			return
		}

		var columnIDs []uint
		collectColumnIDsCache(&columnIDs, column.ID)
		columnIDs = append(columnIDs, column.ID)

		var posts []models.Post
		database.DB.Where("column_id IN ? AND published = ?", columnIDs, true).Preload("Column").Order("id desc").Find(&posts)
		data := gin.H{
			"posts":       posts,
			"columns":     loadColumnsFlat(),
			"column":      column,
			"custom_tags": frontCache.getCustomTags(),
		}
		if column.IsPage {
			data["page_content"] = renderColumnPageContent(column, getSettings(), data)
		}

		renderFront(c, http.StatusOK, data)
	})
}

func collectColumnIDsCache(ids *[]uint, parentID uint) {
	all := frontCache.getColumnsAll()
	for _, c := range all {
		if c.ParentID != nil && *c.ParentID == parentID {
			*ids = append(*ids, c.ID)
			collectColumnIDsCache(ids, c.ID)
		}
	}
}
