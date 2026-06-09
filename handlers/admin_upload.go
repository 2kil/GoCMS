package handlers

import (
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"shuaitesteel.com/cms/config"
)

const maxUploadImageSize = 5 << 20

func UploadImage(c *gin.Context) {
	saveUploadedImage(c, "static", "uploads", "/static/uploads/")
}

func UploadCoverImage(c *gin.Context) {
	month := time.Now().Format("200601")
	saveUploadedImage(c, "upload", month, "/upload/"+month+"/")
}

func saveUploadedImage(c *gin.Context, baseDir, subDir, urlPrefix string) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择图片文件"})
		return
	}
	if file.Size > maxUploadImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片不能超过 5MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedImageExt(ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg、jpeg、png、gif、webp 图片"})
		return
	}
	if !isAllowedImageContentType(file) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件内容不是有效图片"})
		return
	}

	uploadDir := config.ResolvePath(filepath.Join(baseDir, subDir))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	name := uuid.New().String() + ext
	dst := filepath.Join(uploadDir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": urlPrefix + name})
}

func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func isAllowedImageContentType(file *multipart.FileHeader) bool {
	src, err := file.Open()
	if err != nil {
		return false
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	switch http.DetectContentType(buf[:n]) {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}
