package handlers

import (
	"fmt"
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
	month := time.Now().Format("200601")
	saveUploadedImage(c, "upload", month, "/upload/"+month+"/")
}

func UploadCoverImage(c *gin.Context) {
	month := time.Now().Format("200601")
	saveUploadedImage(c, "upload", month, "/upload/"+month+"/")
}

func UploadImages(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}
	files := form.File["images"]
	if len(files) == 0 {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}

	month := time.Now().Format("200601")
	for _, file := range files {
		if _, err := saveUploadedImageFile(c, file, "upload", month, "/upload/"+month+"/"); err != nil {
			continue
		}
	}
	c.Redirect(http.StatusFound, "/adm1n/uploads")
}

func saveUploadedImage(c *gin.Context, baseDir, subDir, urlPrefix string) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择图片文件"})
		return
	}
	url, err := saveUploadedImageFile(c, file, baseDir, subDir, urlPrefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func saveUploadedImageFile(c *gin.Context, file *multipart.FileHeader, baseDir, subDir, urlPrefix string) (string, error) {
	if file.Size > maxUploadImageSize {
		return "", fmt.Errorf("图片不能超过 5MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedImageExt(ext) {
		return "", fmt.Errorf("仅支持 jpg、jpeg、png、gif、webp 图片")
	}
	if !isAllowedImageContentType(file) {
		return "", fmt.Errorf("文件内容不是有效图片")
	}

	uploadDir := config.ResolvePath(filepath.Join(baseDir, subDir))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败")
	}

	name := uuid.New().String() + ext
	dst := filepath.Join(uploadDir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", fmt.Errorf("保存图片失败")
	}

	return urlPrefix + name, nil
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
