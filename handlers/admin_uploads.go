package handlers

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

type uploadImageItem struct {
	Name      string
	URL       string
	Month     string
	Date      string
	Size      string
	SizeBytes int64
}

func ListUploads(c *gin.Context) {
	items := loadUploadedImages()
	c.HTML(http.StatusOK, "uploads.html", gin.H{
		"images":   items,
		"count":    len(items),
		"error":    c.Query("error"),
		"success":  c.Query("success"),
		"nickname": c.MustGet("nickname"),
	})
}

func RenameUpload(c *gin.Context) {
	month := c.PostForm("month")
	name := c.PostForm("name")
	newName := c.PostForm("new_name")
	oldPath, err := uploadImagePath(month, name)
	if err != nil {
		redirectUploads(c, "error", "图片路径无效")
		return
	}
	oldURL, err := UploadImageURL(month, name)
	if err != nil {
		redirectUploads(c, "error", "图片地址无效")
		return
	}

	newName = normalizeUploadFileName(newName, filepath.Ext(name))
	if newName == "" || !isAllowedImageExt(strings.ToLower(filepath.Ext(newName))) {
		redirectUploads(c, "error", "新文件名无效")
		return
	}
	newPath, err := uploadImagePath(month, newName)
	if err != nil {
		redirectUploads(c, "error", "新图片路径无效")
		return
	}
	newURL, err := UploadImageURL(month, newName)
	if err != nil {
		redirectUploads(c, "error", "新图片地址无效")
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		redirectUploads(c, "error", "目标文件已存在")
		return
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		redirectUploads(c, "error", "图片重命名失败")
		return
	}
	if err := ReplaceUploadReferences(oldURL, newURL); err != nil {
		if revertErr := os.Rename(newPath, oldPath); revertErr != nil {
			log.Printf("图片重命名失败后恢复文件失败: %v", revertErr)
		}
		redirectUploads(c, "error", "更新文章图片引用失败: "+err.Error())
		return
	}
	InvalidateCache()
	RequestGenerateStatic()
	redirectUploads(c, "success", "图片已重命名")
}

func DeleteUpload(c *gin.Context) {
	month := c.PostForm("month")
	name := c.PostForm("name")
	path, err := uploadImagePath(month, name)
	if err != nil {
		redirectUploads(c, "error", "图片路径无效")
		return
	}
	imageURL, err := UploadImageURL(month, name)
	if err != nil {
		redirectUploads(c, "error", "图片地址无效")
		return
	}
	refCount, err := CountUploadReferences(imageURL)
	if err != nil {
		redirectUploads(c, "error", "检查图片引用失败")
		return
	}
	if refCount > 0 {
		redirectUploads(c, "error", "图片正在被文章引用，不能删除")
		return
	}
	if err := os.Remove(path); err != nil {
		redirectUploads(c, "error", "图片删除失败")
		return
	}
	redirectUploads(c, "success", "图片已删除")
}

func DeleteUploads(c *gin.Context) {
	deleted := 0
	skipped := 0
	failed := 0
	for _, item := range c.PostFormArray("items") {
		parts := strings.SplitN(item, "\t", 2)
		if len(parts) != 2 {
			failed++
			continue
		}
		path, err := uploadImagePath(parts[0], parts[1])
		if err != nil {
			failed++
			continue
		}
		imageURL, err := UploadImageURL(parts[0], parts[1])
		if err != nil {
			failed++
			continue
		}
		refCount, err := CountUploadReferences(imageURL)
		if err != nil {
			failed++
			continue
		}
		if refCount > 0 {
			skipped++
			continue
		}
		if err := os.Remove(path); err != nil {
			failed++
			continue
		}
		deleted++
	}
	if skipped > 0 || failed > 0 {
		var messages []string
		if deleted > 0 {
			messages = append(messages, "已删除 "+strconv.Itoa(deleted)+" 张")
		}
		if skipped > 0 {
			messages = append(messages, "跳过 "+strconv.Itoa(skipped)+" 张被文章引用的图片")
		}
		if failed > 0 {
			messages = append(messages, strconv.Itoa(failed)+" 张删除失败")
		}
		redirectUploads(c, "error", strings.Join(messages, "，"))
		return
	}
	if deleted == 0 {
		redirectUploads(c, "error", "没有选择可删除的图片")
		return
	}
	redirectUploads(c, "success", "已删除 "+strconv.Itoa(deleted)+" 张图片")
}

func redirectUploads(c *gin.Context, key, message string) {
	if key == "" || message == "" {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}
	c.Redirect(http.StatusFound, "/adm1n/uploads?"+key+"="+url.QueryEscape(message))
}

func loadUploadedImages() []uploadImageItem {
	var items []uploadImageItem

	root := config.ResolvePath("upload")
	entries, err := os.ReadDir(root)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || !isUploadMonthDir(entry.Name()) {
				continue
			}
			month := entry.Name()
			items = appendUploadImageItems(items, filepath.Join(root, month), month, "/upload/"+month+"/")
		}
	}
	items = appendUploadImageItems(items, config.ResolvePath(filepath.Join("static", "uploads")), "static", "/static/uploads/")

	sort.Slice(items, func(i, j int) bool {
		if items[i].Month == items[j].Month {
			return items[i].Name > items[j].Name
		}
		return items[i].Month > items[j].Month
	})
	return items
}

func appendUploadImageItems(items []uploadImageItem, dir, month, urlPrefix string) []uploadImageItem {
	files, err := os.ReadDir(dir)
	if err != nil {
		return items
	}
	for _, file := range files {
		if file.IsDir() || !isAllowedImageExt(strings.ToLower(filepath.Ext(file.Name()))) {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		items = append(items, uploadImageItem{
			Name:      file.Name(),
			URL:       urlPrefix + url.PathEscape(file.Name()),
			Month:     month,
			Date:      info.ModTime().Format("2006-01-02 15:04"),
			Size:      formatFileSize(info.Size()),
			SizeBytes: info.Size(),
		})
	}
	return items
}

func uploadImagePath(month, name string) (string, error) {
	if !isUploadBucket(month) || name == "" || filepath.Base(name) != name || !isAllowedImageExt(strings.ToLower(filepath.Ext(name))) {
		return "", os.ErrInvalid
	}
	if month == "static" {
		return filepath.Join(config.ResolvePath(filepath.Join("static", "uploads")), name), nil
	}
	return filepath.Join(config.ResolvePath("upload"), month, name), nil
}

func UploadImageURL(month, name string) (string, error) {
	if !isUploadBucket(month) || name == "" || filepath.Base(name) != name || !isAllowedImageExt(strings.ToLower(filepath.Ext(name))) {
		return "", os.ErrInvalid
	}
	if month == "static" {
		return "/static/uploads/" + url.PathEscape(name), nil
	}
	return "/upload/" + month + "/" + url.PathEscape(name), nil
}

func CountUploadReferences(imageURL string) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Post{}).
		Where("cover_image = ? OR content LIKE ? ESCAPE '\\'", imageURL, "%"+escapeUploadLike(imageURL)+"%").
		Count(&count).Error
	return count, err
}

func ReplaceUploadReferences(oldURL, newURL string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var posts []models.Post
		if err := tx.Where("cover_image = ? OR content LIKE ? ESCAPE '\\'", oldURL, "%"+escapeUploadLike(oldURL)+"%").Find(&posts).Error; err != nil {
			return err
		}
		for _, post := range posts {
			updates := map[string]interface{}{}
			if post.CoverImage == oldURL {
				updates["cover_image"] = newURL
			}
			if strings.Contains(post.Content, oldURL) {
				updates["content"] = strings.ReplaceAll(post.Content, oldURL, newURL)
			}
			if len(updates) == 0 {
				continue
			}
			if err := tx.Model(&models.Post{}).Where("id = ?", post.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func escapeUploadLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}

func isUploadBucket(name string) bool {
	return name == "static" || isUploadMonthDir(name)
}

func normalizeUploadFileName(name, fallbackExt string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if ext == "" {
		ext = strings.ToLower(fallbackExt)
	}
	base = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, base)
	base = strings.Trim(base, "_")
	if base == "" {
		return ""
	}
	return base + ext
}

func isUploadMonthDir(name string) bool {
	if len(name) != 6 {
		return false
	}
	if _, err := time.Parse("200601", name); err != nil {
		return false
	}
	return true
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return strconv.FormatInt(size, 10) + " B"
	}
	if size < 1024*1024 {
		return strconv.FormatFloat(float64(size)/1024, 'f', 1, 64) + " KB"
	}
	return strconv.FormatFloat(float64(size)/(1024*1024), 'f', 1, 64) + " MB"
}
