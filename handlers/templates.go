package handlers

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"

	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/models"
)

const defaultFrontTemplate = "index"

func availableFrontTemplates() []string {
	entries, err := os.ReadDir(config.TemplateDir())
	if err != nil {
		return []string{defaultFrontTemplate}
	}

	templates := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "admin" {
			continue
		}
		if _, err := os.Stat(filepath.Join(config.TemplateDir(), entry.Name(), "index.html")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(config.TemplateDir(), entry.Name(), "layout.html")); err != nil {
			continue
		}
		templates = append(templates, entry.Name())
	}
	sort.Strings(templates)
	if len(templates) == 0 {
		return []string{defaultFrontTemplate}
	}
	return templates
}

func selectedFrontTemplate(settings map[string]string) string {
	selected := settings["front_template"]
	available := availableFrontTemplates()
	for _, name := range available {
		if name == selected {
			return selected
		}
	}
	for _, name := range available {
		if name == defaultFrontTemplate {
			return defaultFrontTemplate
		}
	}
	if len(available) > 0 {
		return available[0]
	}
	return defaultFrontTemplate
}

func frontTemplates(settings map[string]string) *template.Template {
	tpl := template.New("").Funcs(templateFuncs())
	tpl, err := tpl.ParseGlob(config.FrontTemplateGlob(selectedFrontTemplate(settings)))
	if err != nil {
		log.Printf("加载前台模板失败: %v", err)
		return nil
	}
	return tpl
}

func frontTemplatePageFiles(settings map[string]string) []string {
	dir := filepath.Join(config.TemplateDir(), selectedFrontTemplate(settings))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".html" || name == "index.html" || name == "layout.html" {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)
	return files
}

func renderFrontTemplateFile(settings map[string]string, fileName string, data interface{}) template.HTML {
	if fileName == "" || filepath.Base(fileName) != fileName || filepath.Ext(fileName) != ".html" {
		return ""
	}
	fullPath := filepath.Join(config.TemplateDir(), selectedFrontTemplate(settings), fileName)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("加载单页模板文件失败 %s: %v", fileName, err)
		}
		return ""
	}
	tpl, err := template.New(fileName).Funcs(templateFuncs()).Parse(string(content))
	if err != nil {
		log.Printf("解析单页模板文件失败 %s: %v", fileName, err)
		return ""
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		log.Printf("渲染单页模板文件失败 %s: %v", fileName, err)
		return ""
	}
	return template.HTML(buf.String())
}

func renderColumnPageContent(column models.Column, settings map[string]string, data map[string]interface{}) template.HTML {
	if column.PageTemplate != "" {
		if rendered := renderFrontTemplateFile(settings, column.PageTemplate, data); rendered != "" {
			return rendered
		}
	}
	return renderTemplateString(column.Content, data)
}
