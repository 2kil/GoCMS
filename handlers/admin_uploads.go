package handlers

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"shuaitesteel.com/cms/config"
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
		"nickname": c.MustGet("nickname"),
	})
}

func RenameUpload(c *gin.Context) {
	month := c.PostForm("month")
	name := c.PostForm("name")
	newName := c.PostForm("new_name")
	oldPath, err := uploadImagePath(month, name)
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}

	newName = normalizeUploadFileName(newName, filepath.Ext(name))
	if newName == "" || !isAllowedImageExt(strings.ToLower(filepath.Ext(newName))) {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}
	newPath, err := uploadImagePath(month, newName)
	if err != nil {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		c.Redirect(http.StatusFound, "/adm1n/uploads")
		return
	}
	os.Rename(oldPath, newPath)
	c.Redirect(http.StatusFound, "/adm1n/uploads")
}

func DeleteUpload(c *gin.Context) {
	path, err := uploadImagePath(c.PostForm("month"), c.PostForm("name"))
	if err == nil {
		os.Remove(path)
	}
	c.Redirect(http.StatusFound, "/adm1n/uploads")
}

func loadUploadedImages() []uploadImageItem {
	root := config.ResolvePath("upload")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	var items []uploadImageItem
	for _, entry := range entries {
		if !entry.IsDir() || !isUploadMonthDir(entry.Name()) {
			continue
		}
		month := entry.Name()
		dir := filepath.Join(root, month)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
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
				URL:       "/upload/" + month + "/" + url.PathEscape(file.Name()),
				Month:     month,
				Date:      info.ModTime().Format("2006-01-02 15:04"),
				Size:      formatFileSize(info.Size()),
				SizeBytes: info.Size(),
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Month == items[j].Month {
			return items[i].Name > items[j].Name
		}
		return items[i].Month > items[j].Month
	})
	return items
}

func uploadImagePath(month, name string) (string, error) {
	if !isUploadMonthDir(month) || name == "" || filepath.Base(name) != name || !isAllowedImageExt(strings.ToLower(filepath.Ext(name))) {
		return "", os.ErrInvalid
	}
	return filepath.Join(config.ResolvePath("upload"), month, name), nil
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
