package handlers

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/models"
)

func staticDir() string {
	return config.Cfg.Static.Dir
}

func writeStaticFile(tpl *template.Template, filename string, data map[string]interface{}) {
	path := filepath.Join(staticDir(), filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Printf("静态生成: 创建目录失败 %s: %v", filepath.Dir(path), err)
		return
	}
	f, err := os.Create(path)
	if err != nil {
		log.Printf("静态生成: 创建文件失败 %s: %v", path, err)
		return
	}
	defer f.Close()
	if err := tpl.ExecuteTemplate(f, "index.html", data); err != nil {
		log.Printf("静态生成: 渲染失败 %s: %v", path, err)
	}
}

func cleanGeneratedStatic() {
	dir := staticDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			log.Printf("静态生成: 清理旧文件失败 %s: %v", entry.Name(), err)
		}
	}
}

func copyFrontTemplateAssets(templateName string) {
	srcDir := filepath.Join(config.TemplateDir(), templateName)
	dstDir := staticDir()
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		log.Printf("静态生成: 创建资源目录失败 %s: %v", dstDir, err)
		return
	}

	if err := filepath.WalkDir(srcDir, func(srcPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, srcPath)
		if err != nil || rel == "." {
			return err
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".html") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)
		if entry.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	}); err != nil {
		log.Printf("静态生成: 复制模板资源失败: %v", err)
	}
}

func GenerateStatic() {
	if !config.Cfg.Static.Enable {
		return
	}
	settings := frontCache.getSettings()
	tpl := frontTemplates(settings)
	if tpl == nil {
		return
	}

	dir := staticDir()
	cleanGeneratedStatic()
	os.MkdirAll(filepath.Join(dir, "post"), 0755)
	os.MkdirAll(filepath.Join(dir, "column"), 0755)
	copyFrontTemplateAssets(selectedFrontTemplate(settings))

	customTags := frontCache.getCustomTags()
	columns := frontCache.getColumns()

	var posts []models.Post
	database.DB.Where("published = ?", true).Preload("Column").Preload("User").Order("id desc").Find(&posts)

	writeStaticFile(tpl, "index.html", map[string]interface{}{
		"posts":       posts,
		"columns":     columns,
		"settings":    settings,
		"custom_tags": customTags,
	})

	for _, p := range posts {
		writeStaticFile(tpl, filepath.Join("post", p.Slug+".html"), map[string]interface{}{
			"post":        p,
			"columns":     columns,
			"settings":    settings,
			"custom_tags": customTags,
		})
	}

	for _, col := range frontCache.getColumnsAll() {
		genColumnStatic(tpl, col, columns, settings)
	}

	log.Printf("静态页面已全量生成到 %s/", dir)
}

func RemoveStaticPost(slug string) {
	p := filepath.Join(staticDir(), "post", slug+".html")
	os.Remove(p)
}

func GenerateStaticIndex() {
	if !config.Cfg.Static.Enable {
		return
	}
	settings := frontCache.getSettings()
	tpl := frontTemplates(settings)
	if tpl == nil {
		return
	}
	os.MkdirAll(filepath.Join(staticDir(), "post"), 0755)
	os.MkdirAll(filepath.Join(staticDir(), "column"), 0755)

	customTags := frontCache.getCustomTags()
	columns := frontCache.getColumns()

	var posts []models.Post
	database.DB.Where("published = ?", true).Preload("Column").Preload("User").Order("id desc").Find(&posts)

	writeStaticFile(tpl, "index.html", map[string]interface{}{
		"posts":       posts,
		"columns":     columns,
		"settings":    settings,
		"custom_tags": customTags,
	})
}

func GenerateStaticPost(slug string) {
	if !config.Cfg.Static.Enable {
		return
	}
	settings := frontCache.getSettings()
	tpl := frontTemplates(settings)
	if tpl == nil {
		return
	}
	os.MkdirAll(filepath.Join(staticDir(), "post"), 0755)

	customTags := frontCache.getCustomTags()
	columns := frontCache.getColumns()

	var post models.Post
	database.DB.Where("slug = ? AND published = ?", slug, true).Preload("Column").Preload("User").First(&post)
	if post.ID == 0 {
		RemoveStaticPost(slug)
		return
	}

	writeStaticFile(tpl, filepath.Join("post", slug+".html"), map[string]interface{}{
		"post":        post,
		"columns":     columns,
		"settings":    settings,
		"custom_tags": customTags,
	})
}

func GenerateStaticColumnByID(columnID uint) {
	if !config.Cfg.Static.Enable {
		return
	}
	settings := frontCache.getSettings()
	tpl := frontTemplates(settings)
	if tpl == nil {
		return
	}
	os.MkdirAll(filepath.Join(staticDir(), "column"), 0755)

	columns := frontCache.getColumns()

	var col models.Column
	database.DB.First(&col, columnID)
	if col.ID == 0 {
		return
	}

	genColumnStatic(tpl, col, columns, settings)
}

func genColumnStatic(tpl *template.Template, col models.Column, columns []models.Column, settings map[string]string) {
	var columnIDs []uint
	collectColumnIDsCache(&columnIDs, col.ID)
	columnIDs = append(columnIDs, col.ID)

	var colPosts []models.Post
	database.DB.Where("column_id IN ? AND published = ?", columnIDs, true).Preload("Column").Order("id desc").Find(&colPosts)
	data := map[string]interface{}{
		"posts":       colPosts,
		"columns":     columns,
		"column":      col,
		"settings":    settings,
		"custom_tags": frontCache.getCustomTags(),
	}
	if col.IsPage {
		data["page_content"] = renderColumnPageContent(col, settings, data)
	}

	writeStaticFile(tpl, filepath.Join("column", col.Slug+".html"), data)
}
