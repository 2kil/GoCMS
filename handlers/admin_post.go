package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func Dashboard(c *gin.Context) {
	var postCount int64
	var publishedCount int64
	var draftCount int64
	var columnCount int64
	var tagCount int64
	database.DB.Model(&models.Post{}).Count(&postCount)
	database.DB.Model(&models.Post{}).Where("published = ?", true).Count(&publishedCount)
	database.DB.Model(&models.Post{}).Where("published = ?", false).Count(&draftCount)
	database.DB.Model(&models.Column{}).Count(&columnCount)
	database.DB.Model(&models.CustomTag{}).Count(&tagCount)

	lastLoginIP := "暂无记录"
	lastLoginAt := "暂无记录"
	session := sessions.Default(c)
	if v, ok := session.Get("last_login_ip").(string); ok && v != "" {
		lastLoginIP = v
	}
	if v, ok := session.Get("last_login_at").(string); ok && v != "" {
		lastLoginAt = v
	}
	if lastLoginIP == "暂无记录" || lastLoginAt == "暂无记录" {
		var user models.User
		if database.DB.First(&user, c.GetUint("user_id")).Error == nil {
			if lastLoginIP == "暂无记录" && user.LastLoginIP != "" {
				lastLoginIP = user.LastLoginIP
			}
			if lastLoginAt == "暂无记录" && user.LastLoginAt != nil {
				lastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
			}
		}
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"postCount":      postCount,
		"publishedCount": publishedCount,
		"draftCount":     draftCount,
		"columnCount":    columnCount,
		"tagCount":       tagCount,
		"uploadCount":    len(loadUploadedImages()),
		"lastLoginIP":    lastLoginIP,
		"lastLoginAt":    lastLoginAt,
		"nickname":       c.MustGet("nickname"),
	})
}

func ListPosts(c *gin.Context) {
	var posts []models.Post
	database.DB.Preload("Column").Preload("User").Order("id desc").Find(&posts)

	c.HTML(http.StatusOK, "posts.html", gin.H{
		"posts":    posts,
		"nickname": c.MustGet("nickname"),
	})
}

func ShowPostEdit(c *gin.Context) {
	idStr := c.Param("id")
	var post models.Post
	var columns []models.Column

	database.DB.Order("sort_order asc").Find(&columns)

	if idStr != "" && idStr != "new" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || database.DB.First(&post, id).Error != nil {
			c.Redirect(http.StatusFound, "/adm1n/posts")
			return
		}
	}

	var columnID uint
	if post.ColumnID != nil {
		columnID = *post.ColumnID
	}

	c.HTML(http.StatusOK, "post_edit.html", gin.H{
		"post":           post,
		"columns":        columns,
		"uploadedImages": loadUploadedImages(),
		"column_id":      columnID,
		"content_format": normalizeContentFormat(post.ContentFormat),
		"nickname":       c.MustGet("nickname"),
	})
}

func SavePost(c *gin.Context) {
	idStr := c.Param("id")

	title := c.PostForm("title")
	slug := c.PostForm("slug")
	summary := c.PostForm("summary")
	content := c.PostForm("content")
	contentFormat := normalizeContentFormat(c.PostForm("content_format"))
	coverImage := c.PostForm("cover_image")
	published := c.PostForm("published") == "on"
	scheduled := c.PostForm("scheduled") == "on"
	scheduledAtStr := c.PostForm("scheduled_at")
	columnIDStr := c.PostForm("column_id")

	if slug == "" {
		slug = uuid.New().String()
	}

	userID := c.GetUint("user_id")

	var post models.Post

	if idStr != "" && idStr != "new" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || database.DB.First(&post, id).Error != nil {
			c.Redirect(http.StatusFound, "/adm1n/posts")
			return
		}
	}

	post.Title = title
	post.Slug = slug
	post.Summary = summary
	post.Content = content
	post.ContentFormat = contentFormat
	post.CoverImage = coverImage
	post.Published = published
	post.ScheduledAt = nil
	if !published && scheduled && scheduledAtStr != "" {
		if scheduledAt, err := time.ParseInLocation("2006-01-02T15:04", scheduledAtStr, time.Local); err == nil {
			post.ScheduledAt = &scheduledAt
		}
	}
	post.UserID = userID

	if columnIDStr != "" {
		cid, err := strconv.ParseUint(columnIDStr, 10, 64)
		if err == nil {
			cidUint := uint(cid)
			post.ColumnID = &cidUint
		}
	} else {
		post.ColumnID = nil
	}

	var dbErr error
	if post.ID != 0 {
		dbErr = database.DB.Save(&post).Error
	} else {
		dbErr = database.DB.Create(&post).Error
	}
	if dbErr != nil {
		var columns []models.Column
		database.DB.Order("sort_order asc").Find(&columns)
		var columnID uint
		if post.ColumnID != nil {
			columnID = *post.ColumnID
		}
		c.HTML(http.StatusOK, "post_edit.html", gin.H{
			"error":          "保存失败: " + dbErr.Error(),
			"post":           post,
			"columns":        columns,
			"uploadedImages": loadUploadedImages(),
			"column_id":      columnID,
			"content_format": normalizeContentFormat(post.ContentFormat),
			"nickname":       c.MustGet("nickname"),
		})
		return
	}

	InvalidateCache()
	GenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/posts")
}

func normalizeContentFormat(format string) string {
	if format == "html" {
		return "html"
	}
	return "markdown"
}

func TogglePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/posts")
		return
	}
	var post models.Post
	if database.DB.First(&post, id).Error != nil {
		c.Redirect(http.StatusFound, "/adm1n/posts")
		return
	}
	if post.ID != 0 {
		post.Published = !post.Published
		if post.Published {
			post.ScheduledAt = nil
		}
		database.DB.Model(&post).Updates(map[string]interface{}{
			"published":    post.Published,
			"scheduled_at": post.ScheduledAt,
		})
	}
	InvalidateCache()
	GenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/posts")
}

func DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/posts")
		return
	}
	var post models.Post
	if database.DB.First(&post, id).Error != nil {
		c.Redirect(http.StatusFound, "/adm1n/posts")
		return
	}
	database.DB.Delete(&post)
	InvalidateCache()
	GenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/posts")
}
