package handlers

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

var customTagKeyPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

func ListTags(c *gin.Context) {
	var tags []models.CustomTag
	database.DB.Order("id desc").Find(&tags)

	c.HTML(http.StatusOK, "tags.html", gin.H{
		"tags":     tags,
		"nickname": c.MustGet("nickname"),
	})
}

func ShowTagEdit(c *gin.Context) {
	idStr := c.Param("id")
	var tag models.CustomTag

	if idStr != "" && idStr != "new" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || database.DB.First(&tag, id).Error != nil {
			c.Redirect(http.StatusFound, "/adm1n/tags")
			return
		}
	}

	c.HTML(http.StatusOK, "tag_edit.html", gin.H{
		"tag":      tag,
		"nickname": c.MustGet("nickname"),
	})
}

func SaveTag(c *gin.Context) {
	idStr := c.Param("id")
	var tag models.CustomTag

	if idStr != "" && idStr != "new" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || database.DB.First(&tag, id).Error != nil {
			c.Redirect(http.StatusFound, "/adm1n/tags")
			return
		}
	}

	tag.Name = c.PostForm("name")
	tag.Key = c.PostForm("key")
	tag.Value = c.PostForm("value")

	if tag.Name == "" || !customTagKeyPattern.MatchString(tag.Key) {
		c.HTML(http.StatusOK, "tag_edit.html", gin.H{
			"error":    "标签名称不能为空，标签变量只能使用字母、数字、下划线且以字母开头",
			"tag":      tag,
			"nickname": c.MustGet("nickname"),
		})
		return
	}

	var dbErr error
	if tag.ID != 0 {
		dbErr = database.DB.Save(&tag).Error
	} else {
		dbErr = database.DB.Create(&tag).Error
	}
	if dbErr != nil {
		c.HTML(http.StatusOK, "tag_edit.html", gin.H{
			"error":    "保存失败: " + dbErr.Error(),
			"tag":      tag,
			"nickname": c.MustGet("nickname"),
		})
		return
	}

	InvalidateCache()
	RequestGenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/tags")
}

func DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/tags")
		return
	}

	var tag models.CustomTag
	if database.DB.First(&tag, id).Error != nil {
		c.Redirect(http.StatusFound, "/adm1n/tags")
		return
	}

	database.DB.Delete(&tag)
	InvalidateCache()
	RequestGenerateStatic()
	c.Redirect(http.StatusFound, "/adm1n/tags")
}
