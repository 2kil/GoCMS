package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"shuaitesteel.com/cms/config"
	"shuaitesteel.com/cms/database"
	"shuaitesteel.com/cms/handlers"
	"shuaitesteel.com/cms/models"
)

type cliOptions map[string]string

func runCLI(args []string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(2)
	}
	if args[0] == "reset-admin" {
		if len(args) != 3 {
			failUsage("用法: cms reset-admin <username> <password>")
		}
		resetAdminPassword(args[1], args[2])
		return
	}
	if len(args) < 2 || isHelpArg(args[1]) {
		printCommandHelp(args[0])
		return
	}

	switch args[0] {
	case "user":
		runUserCLI(args[1], parseOptions(args[2:]))
	case "post":
		runPostCLI(args[1], parseOptions(args[2:]))
	case "column":
		runColumnCLI(args[1], parseOptions(args[2:]))
	case "tag":
		runTagCLI(args[1], parseOptions(args[2:]))
	case "setting":
		runSettingCLI(args[1], parseOptions(args[2:]))
	case "upload":
		runUploadCLI(args[1], parseOptions(args[2:]))
	case "db":
		runDBCLI(args[1], parseOptions(args[2:]))
	}
}

func parseOptions(args []string) cliOptions {
	opts := make(cliOptions)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			failUsage("无效参数: " + arg)
		}
		key := strings.TrimPrefix(arg, "--")
		if key == "" {
			failUsage("无效参数: " + arg)
		}
		if idx := strings.IndexByte(key, '='); idx >= 0 {
			opts[key[:idx]] = key[idx+1:]
			continue
		}
		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
			opts[key] = "true"
			continue
		}
		opts[key] = args[i+1]
		i++
	}
	return opts
}

func (o cliOptions) get(key string) string { return strings.TrimSpace(o[key]) }

func (o cliOptions) bool(key string) bool {
	switch strings.ToLower(o[key]) {
	case "1", "true", "yes", "on", "发布", "是":
		return true
	default:
		return false
	}
}

func (o cliOptions) uint(key string) (uint, bool) {
	v := o.get(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		fail("参数 %s 必须是数字: %s", key, v)
	}
	return uint(n), true
}

func (o cliOptions) int(key string) (int, bool) {
	v := o.get(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		fail("参数 %s 必须是数字: %s", key, v)
	}
	return n, true
}

func (o cliOptions) time(key string) *time.Time {
	v := o.get(key)
	if v == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return &t
		}
	}
	fail("参数 %s 时间格式无效: %s", key, v)
	return nil
}

func isHelpArg(arg string) bool { return arg == "help" || arg == "-h" || arg == "--help" }

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func failUsage(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}

func runUserCLI(action string, opts cliOptions) {
	switch action {
	case "list":
		var users []models.User
		mustDB(database.DB.Order("id asc").Find(&users).Error)
		fmt.Println("ID\t用户名\t昵称\t管理员")
		for _, u := range users {
			fmt.Printf("%d\t%s\t%s\t%t\n", u.ID, u.Username, u.Nickname, u.IsAdmin)
		}
	case "create":
		username := required(opts, "username")
		password := required(opts, "password")
		nickname := opts.get("nickname")
		if nickname == "" {
			nickname = username
		}
		hashed := hashPassword(password)
		user := models.User{Username: username, Password: hashed, Nickname: nickname, IsAdmin: opts.bool("admin")}
		mustDB(database.DB.Create(&user).Error)
		fmt.Printf("用户已创建: id=%d username=%s\n", user.ID, user.Username)
	case "password":
		resetAdminPassword(required(opts, "username"), required(opts, "password"))
	case "delete":
		id := requiredID(opts)
		mustDB(database.DB.Delete(&models.User{}, id).Error)
		fmt.Printf("用户已删除: id=%d\n", id)
	default:
		printCommandHelp("user")
		os.Exit(2)
	}
}

func runPostCLI(action string, opts cliOptions) {
	switch action {
	case "list":
		var posts []models.Post
		mustDB(database.DB.Preload("Column").Order("id desc").Find(&posts).Error)
		fmt.Println("ID\tSlug\t标题\t发布\t栏目ID")
		for _, p := range posts {
			columnID := ""
			if p.ColumnID != nil {
				columnID = strconv.FormatUint(uint64(*p.ColumnID), 10)
			}
			fmt.Printf("%d\t%s\t%s\t%t\t%s\n", p.ID, p.Slug, p.Title, p.Published, columnID)
		}
	case "get":
		var post models.Post
		mustFind(findPost(opts, &post))
		fmt.Printf("ID: %d\n标题: %s\nSlug: %s\n摘要: %s\n格式: %s\n发布: %t\n内容:\n%s\n", post.ID, post.Title, post.Slug, post.Summary, normalizeCLIContentFormat(post.ContentFormat), post.Published, post.Content)
	case "save":
		post := models.Post{}
		if id, ok := opts.uint("id"); ok {
			mustFind(database.DB.First(&post, id).Error)
		}
		applyPostOptions(&post, opts)
		if post.ID == 0 && post.UserID == 0 {
			post.UserID = firstUserID()
		}
		if post.ID == 0 {
			mustDB(database.DB.Create(&post).Error)
		} else {
			mustDB(database.DB.Save(&post).Error)
		}
		afterContentChange()
		fmt.Printf("文章已保存: id=%d slug=%s\n", post.ID, post.Slug)
	case "delete":
		var post models.Post
		mustFind(findPost(opts, &post))
		mustDB(database.DB.Delete(&post).Error)
		afterContentChange()
		fmt.Printf("文章已删除: id=%d\n", post.ID)
	case "publish", "unpublish":
		var post models.Post
		mustFind(findPost(opts, &post))
		post.Published = action == "publish"
		if post.Published {
			post.ScheduledAt = nil
		}
		mustDB(database.DB.Model(&post).Updates(map[string]interface{}{
			"published":    post.Published,
			"scheduled_at": post.ScheduledAt,
		}).Error)
		afterContentChange()
		fmt.Printf("文章发布状态已更新: id=%d published=%t\n", post.ID, post.Published)
	default:
		printCommandHelp("post")
		os.Exit(2)
	}
}

func applyPostOptions(post *models.Post, opts cliOptions) {
	if v := opts.get("title"); v != "" {
		post.Title = v
	}
	if v := opts.get("slug"); v != "" {
		if !handlers.IsValidSlug(v) {
			fail("文章 slug 只能使用字母、数字、横线和下划线，且必须以字母或数字开头")
		}
		post.Slug = v
	}
	if _, ok := opts["summary"]; ok {
		post.Summary = opts.get("summary")
	}
	if _, ok := opts["content"]; ok {
		post.Content = opts.get("content")
	}
	if v := opts.get("content-format"); v != "" {
		post.ContentFormat = normalizeCLIContentFormat(v)
	}
	if _, ok := opts["cover"]; ok {
		post.CoverImage = opts.get("cover")
	}
	if v, ok := opts.uint("column-id"); ok {
		post.ColumnID = &v
	}
	if _, ok := opts["no-column"]; ok {
		post.ColumnID = nil
	}
	if _, ok := opts["published"]; ok {
		post.Published = opts.bool("published")
		if post.Published {
			post.ScheduledAt = nil
		}
	}
	if scheduledAt := opts.time("scheduled-at"); scheduledAt != nil {
		post.Published = false
		post.ScheduledAt = scheduledAt
	}
	if post.Title == "" || post.Slug == "" {
		fail("文章 title 和 slug 不能为空")
	}
}

func runColumnCLI(action string, opts cliOptions) {
	switch action {
	case "list":
		var columns []models.Column
		mustDB(database.DB.Order("sort_order asc, id asc").Find(&columns).Error)
		fmt.Println("ID\tParentID\tSort\tSlug\t名称\t单页")
		for _, c := range columns {
			parentID := ""
			if c.ParentID != nil {
				parentID = strconv.FormatUint(uint64(*c.ParentID), 10)
			}
			fmt.Printf("%d\t%s\t%d\t%s\t%s\t%t\n", c.ID, parentID, c.SortOrder, c.Slug, c.Name, c.IsPage)
		}
	case "get":
		var column models.Column
		mustFind(findColumn(opts, &column))
		fmt.Printf("ID: %d\n名称: %s\nSlug: %s\n单页: %t\n模板: %s\n内容:\n%s\n", column.ID, column.Name, column.Slug, column.IsPage, column.PageTemplate, column.Content)
	case "save":
		column := models.Column{}
		if id, ok := opts.uint("id"); ok {
			mustFind(database.DB.First(&column, id).Error)
		}
		applyColumnOptions(&column, opts)
		if column.ID == 0 {
			mustDB(database.DB.Create(&column).Error)
		} else {
			mustDB(database.DB.Save(&column).Error)
		}
		afterContentChange()
		fmt.Printf("栏目已保存: id=%d slug=%s\n", column.ID, column.Slug)
	case "delete":
		var column models.Column
		mustFind(findColumn(opts, &column))
		mustDB(database.DB.Model(&models.Column{}).Where("parent_id = ?", column.ID).Update("parent_id", nil).Error)
		mustDB(database.DB.Delete(&column).Error)
		afterContentChange()
		fmt.Printf("栏目已删除: id=%d\n", column.ID)
	default:
		printCommandHelp("column")
		os.Exit(2)
	}
}

func applyColumnOptions(column *models.Column, opts cliOptions) {
	if v := opts.get("name"); v != "" {
		column.Name = v
	}
	if v := opts.get("slug"); v != "" {
		if !handlers.IsValidSlug(v) {
			fail("栏目 slug 只能使用字母、数字、横线和下划线，且必须以字母或数字开头")
		}
		column.Slug = v
	}
	if _, ok := opts["is-page"]; ok {
		column.IsPage = opts.bool("is-page")
	}
	if v := opts.get("page-template"); v != "" {
		column.PageTemplate = v
	}
	if v := opts.get("content"); v != "" {
		column.Content = v
	}
	if v, ok := opts.int("sort"); ok {
		column.SortOrder = v
	}
	if v, ok := opts.uint("parent-id"); ok {
		if column.ID != 0 && (v == column.ID || isDescendantColumnCLI(v, column.ID)) {
			fail("父栏目不能选择当前栏目或其子栏目")
		}
		column.ParentID = &v
	}
	if _, ok := opts["no-parent"]; ok {
		column.ParentID = nil
	}
	if column.Name == "" || column.Slug == "" {
		fail("栏目 name 和 slug 不能为空")
	}
}

func runTagCLI(action string, opts cliOptions) {
	switch action {
	case "list":
		var tags []models.CustomTag
		mustDB(database.DB.Order("id desc").Find(&tags).Error)
		fmt.Println("ID\tKey\t名称\t内容")
		for _, tag := range tags {
			fmt.Printf("%d\t%s\t%s\t%s\n", tag.ID, tag.Key, tag.Name, tag.Value)
		}
	case "get":
		key := required(opts, "key")
		var tag models.CustomTag
		mustFind(database.DB.Where("key = ?", key).First(&tag).Error)
		fmt.Printf("ID: %d\nKey: %s\n名称: %s\n内容:\n%s\n", tag.ID, tag.Key, tag.Name, tag.Value)
	case "set":
		key := required(opts, "key")
		tag := models.CustomTag{}
		mustDB(database.DB.Where("key = ?", key).Find(&tag).Error)
		tag.Key = key
		if v := opts.get("name"); v != "" {
			tag.Name = v
		}
		if tag.Name == "" {
			tag.Name = key
		}
		tag.Value = opts.get("value")
		mustDB(database.DB.Save(&tag).Error)
		afterContentChange()
		fmt.Printf("自定义标签已保存: key=%s\n", tag.Key)
	case "delete":
		key := required(opts, "key")
		mustDB(database.DB.Where("key = ?", key).Delete(&models.CustomTag{}).Error)
		afterContentChange()
		fmt.Printf("自定义标签已删除: key=%s\n", key)
	default:
		printCommandHelp("tag")
		os.Exit(2)
	}
}

func runSettingCLI(action string, opts cliOptions) {
	switch action {
	case "list":
		var settings []models.Setting
		mustDB(database.DB.Order("key asc").Find(&settings).Error)
		fmt.Println("Key\tValue")
		for _, s := range settings {
			fmt.Printf("%s\t%s\n", s.Key, s.Value)
		}
	case "get":
		key := required(opts, "key")
		var setting models.Setting
		mustFind(database.DB.Where("key = ?", key).First(&setting).Error)
		fmt.Println(setting.Value)
	case "set":
		key := required(opts, "key")
		setting := models.Setting{}
		mustDB(database.DB.Where("key = ?", key).Find(&setting).Error)
		setting.Key = key
		setting.Value = opts.get("value")
		mustDB(database.DB.Save(&setting).Error)
		afterContentChange()
		fmt.Printf("设置已保存: key=%s\n", key)
	default:
		printCommandHelp("setting")
		os.Exit(2)
	}
}

func runUploadCLI(action string, opts cliOptions) {
	switch action {
	case "list":
		items := loadCLIUploadedImages()
		fmt.Println("Month\tName\tDate\tSize\tURL")
		for _, item := range items {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", item.Month, item.Name, item.Date, item.Size, item.URL)
		}
	case "rename":
		month := required(opts, "month")
		name := required(opts, "name")
		newName := normalizeCLIUploadFileName(required(opts, "new-name"), filepath.Ext(name))
		if newName == "" || !isAllowedCLIImageExt(strings.ToLower(filepath.Ext(newName))) {
			fail("新文件名无效: %s", opts.get("new-name"))
		}
		oldPath, err := cliUploadImagePath(month, name)
		if err != nil {
			fail("图片路径无效")
		}
		newPath, err := cliUploadImagePath(month, newName)
		if err != nil {
			fail("新图片路径无效")
		}
		if _, err := os.Stat(newPath); err == nil {
			fail("目标文件已存在: %s", newName)
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			fail("图片重命名失败: %v", err)
		}
		fmt.Printf("图片已重命名: %s/%s -> %s/%s\n", month, name, month, newName)
	case "delete":
		month := required(opts, "month")
		name := required(opts, "name")
		path, err := cliUploadImagePath(month, name)
		if err != nil {
			fail("图片路径无效")
		}
		if err := os.Remove(path); err != nil {
			fail("图片删除失败: %v", err)
		}
		fmt.Printf("图片已删除: %s/%s\n", month, name)
	default:
		printCommandHelp("upload")
		os.Exit(2)
	}
}

func runDBCLI(action string, opts cliOptions) {
	switch action {
	case "init", "migrate":
		fmt.Println("数据库已初始化并迁移")
	case "schema":
		printDBSchema()
	case "repair":
		repairDB(opts.bool("generate-static"))
	default:
		printCommandHelp("db")
		os.Exit(2)
	}
}

func printDBSchema() {
	var tables []string
	mustDB(database.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&tables).Error)
	for _, table := range tables {
		fmt.Printf("[%s]\n", table)
		var rows []struct {
			CID       int    `gorm:"column:cid"`
			Name      string `gorm:"column:name"`
			Type      string `gorm:"column:type"`
			NotNull   int    `gorm:"column:notnull"`
			DfltValue string `gorm:"column:dflt_value"`
			PK        int    `gorm:"column:pk"`
		}
		mustDB(database.DB.Raw("PRAGMA table_info(" + table + ")").Scan(&rows).Error)
		for _, row := range rows {
			fmt.Printf("  %s\t%s\tnotnull=%d\tpk=%d\tdefault=%s\n", row.Name, row.Type, row.NotNull, row.PK, row.DfltValue)
		}
	}
}

func repairDB(generateStatic bool) {
	var updates int64
	updates += execRepair("UPDATE columns SET sort_order = id WHERE sort_order = 0")
	updates += execRepair("UPDATE columns SET slug = 'column-' || id WHERE slug = '' OR slug IS NULL")
	updates += execRepair("UPDATE posts SET slug = 'post-' || id WHERE slug = '' OR slug IS NULL")
	updates += execRepair("UPDATE posts SET title = slug WHERE title = '' OR title IS NULL")
	updates += execRepair("UPDATE custom_tags SET name = key WHERE name = '' OR name IS NULL")
	updates += execRepair("UPDATE users SET nickname = username WHERE nickname = '' OR nickname IS NULL")
	if generateStatic {
		afterContentChange()
	}
	fmt.Printf("数据库修复完成，影响行数: %d\n", updates)
}

type cliUploadImageItem struct {
	Name      string
	URL       string
	Month     string
	Date      string
	Size      string
	SizeBytes int64
}

func loadCLIUploadedImages() []cliUploadImageItem {
	var items []cliUploadImageItem

	root := config.ResolvePath("upload")
	entries, err := os.ReadDir(root)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || !isCLIUploadMonthDir(entry.Name()) {
				continue
			}
			month := entry.Name()
			items = appendCLIUploadImageItems(items, filepath.Join(root, month), month, "/upload/"+month+"/")
		}
	}
	items = appendCLIUploadImageItems(items, config.ResolvePath(filepath.Join("static", "uploads")), "static", "/static/uploads/")

	sort.Slice(items, func(i, j int) bool {
		if items[i].Month == items[j].Month {
			return items[i].Name > items[j].Name
		}
		return items[i].Month > items[j].Month
	})
	return items
}

func appendCLIUploadImageItems(items []cliUploadImageItem, dir, month, urlPrefix string) []cliUploadImageItem {
	files, err := os.ReadDir(dir)
	if err != nil {
		return items
	}
	for _, file := range files {
		if file.IsDir() || !isAllowedCLIImageExt(strings.ToLower(filepath.Ext(file.Name()))) {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		items = append(items, cliUploadImageItem{
			Name:      file.Name(),
			URL:       urlPrefix + url.PathEscape(file.Name()),
			Month:     month,
			Date:      info.ModTime().Format("2006-01-02 15:04"),
			Size:      formatCLIFileSize(info.Size()),
			SizeBytes: info.Size(),
		})
	}
	return items
}

func cliUploadImagePath(month, name string) (string, error) {
	if !isCLIUploadBucket(month) || name == "" || filepath.Base(name) != name || !isAllowedCLIImageExt(strings.ToLower(filepath.Ext(name))) {
		return "", os.ErrInvalid
	}
	if month == "static" {
		return filepath.Join(config.ResolvePath(filepath.Join("static", "uploads")), name), nil
	}
	return filepath.Join(config.ResolvePath("upload"), month, name), nil
}

func isCLIUploadBucket(name string) bool {
	return name == "static" || isCLIUploadMonthDir(name)
}

func normalizeCLIUploadFileName(name, fallbackExt string) string {
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

func isCLIUploadMonthDir(name string) bool {
	if len(name) != 6 {
		return false
	}
	if _, err := time.Parse("200601", name); err != nil {
		return false
	}
	return true
}

func isAllowedCLIImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func formatCLIFileSize(size int64) string {
	if size < 1024 {
		return strconv.FormatInt(size, 10) + " B"
	}
	if size < 1024*1024 {
		return strconv.FormatFloat(float64(size)/1024, 'f', 1, 64) + " KB"
	}
	return strconv.FormatFloat(float64(size)/(1024*1024), 'f', 1, 64) + " MB"
}

func execRepair(sql string) int64 {
	result := database.DB.Exec(sql)
	mustDB(result.Error)
	return result.RowsAffected
}

func afterContentChange() {
	handlers.InvalidateCache()
	handlers.GenerateStatic()
}

func resetAdminPassword(username, password string) {
	if username == "" || password == "" {
		failUsage("用户名和密码不能为空")
	}
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		fail("用户不存在: %s", username)
	}
	user.Password = hashPassword(password)
	mustDB(database.DB.Save(&user).Error)
	fmt.Printf("管理员密码已重置: %s\n", username)
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fail("密码加密失败: %v", err)
	}
	return string(hashed)
}

func firstUserID() uint {
	var user models.User
	if err := database.DB.Order("id asc").First(&user).Error; err != nil {
		fail("无法找到默认作者，请先创建后台用户")
	}
	return user.ID
}

func normalizeCLIContentFormat(format string) string {
	if format == "html" {
		return "html"
	}
	return "markdown"
}

func findPost(opts cliOptions, post *models.Post) error {
	if id, ok := opts.uint("id"); ok {
		return database.DB.First(post, id).Error
	}
	return database.DB.Where("slug = ?", required(opts, "slug")).First(post).Error
}

func findColumn(opts cliOptions, column *models.Column) error {
	if id, ok := opts.uint("id"); ok {
		return database.DB.First(column, id).Error
	}
	return database.DB.Where("slug = ?", required(opts, "slug")).First(column).Error
}

func isDescendantColumnCLI(candidateID uint, columnID uint) bool {
	currentID := candidateID
	for currentID != 0 {
		if currentID == columnID {
			return true
		}
		var current models.Column
		if err := database.DB.First(&current, currentID).Error; err != nil || current.ParentID == nil {
			return false
		}
		currentID = *current.ParentID
	}
	return false
}

func required(opts cliOptions, key string) string {
	v := opts.get(key)
	if v == "" {
		failUsage("缺少参数 --" + key)
	}
	return v
}

func requiredID(opts cliOptions) uint {
	id, ok := opts.uint("id")
	if !ok {
		failUsage("缺少参数 --id")
	}
	return id
}

func mustDB(err error) {
	if err != nil {
		fail("数据库操作失败: %v", err)
	}
}

func mustFind(err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fail("记录不存在")
	}
	mustDB(err)
}

func printCommandHelp(command string) {
	switch command {
	case "user":
		fmt.Println("user: list | create --username --password [--nickname] [--admin] | password --username --password | delete --id")
	case "post":
		fmt.Println("post: list | get (--id|--slug) | save [--id] --title --slug [--summary] [--content] [--content-format markdown|html] [--cover] [--column-id] [--published true|false] [--scheduled-at] | delete (--id|--slug) | publish (--id|--slug) | unpublish (--id|--slug)")
	case "column":
		fmt.Println("column: list | get (--id|--slug) | save [--id] --name --slug [--parent-id] [--no-parent] [--sort] [--is-page] [--page-template] [--content] | delete (--id|--slug)")
	case "tag":
		fmt.Println("tag: list | get --key | set --key --value [--name] | delete --key")
	case "setting":
		fmt.Println("setting: list | get --key | set --key --value")
	case "upload":
		fmt.Println("upload: list | rename --month --name --new-name | delete --month --name")
	case "db":
		fmt.Println("db: init | migrate | schema | repair [--generate-static]")
	default:
		printUsage()
	}
}
