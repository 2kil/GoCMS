# CMS 项目维护说明

本文档用于记录当前 CMS 的运行方式、后台使用方式、模板变量、自定义标签和直接操作数据库的注意事项。后续维护本项目时，优先遵循本文档。

## 项目定位

这是一个基于 Go、Gin、GORM 和 SQLite 的轻量 CMS。

当前设计是：

- Gin 主要负责后台管理功能。
- 前台页面通过模板生成静态 HTML 文件。
- 生成结果放在 `public/` 目录。
- 前台访问未命中后台路由时，由 `ServePublic` 从 `public/` 读取静态文件。
- 前台模板自带 CSS、JS、图片等资源放在对应模板目录下，静态生成时会复制到 `public/`。
- 后台样式资源放在 `templates/admin/css/`，通过 `/admin-assets/css/...` 访问。
- 编译后的二进制支持 CLI；不带参数的 `cms` 必须继续等同于 `cms serve`，用于兼容现有生产启动方式。
- 本项目已经用于生产环境，后续代码和数据库变更必须优先考虑向前兼容和自动升级。

## 常用命令

运行项目：

```bash
go run .
```

显式启动 Web 服务：

```bash
go run . serve
```

CLI 帮助：

```bash
go run . help
go run . user help
go run . db schema
```

编译和基础检查：

```bash
go test ./...
```

格式化 Go 代码：

```bash
gofmt -w .
```

打包可参考：

```bash
./build.ps1
```

## 默认配置

配置文件：`config.ini`

当前关键配置：

```ini
[server]
port = 8080

[database]
path = cms.db

[static]
enable = true
dir = public
```

默认后台地址：

```text
http://localhost:8080/adm1n/login
```

默认账号：

```text
admin / G0u8NmtXSsFmDwxDCl
```

## 目录说明

关键目录：

- `main.go`：应用入口、路由注册、模板函数注册。
- `cli.go`：CLI 子命令实现，覆盖后台常用维护操作和数据库维护操作。
- `config/`：配置加载和模板路径解析。
- `database/`：数据库初始化、默认数据写入。
- `models/`：数据库模型。
- `handlers/`：后台、前台、静态生成逻辑。
- `middleware/`：后台登录校验。
- `templates/admin/`：后台模板。
- `templates/index/`：前台模板。
- `static/`：预留的公共静态资源目录。
- `templates/admin/css/`：后台样式资源。
- `templates/index/css/`：默认前台模板样式资源。
- `public/`：生成后的前台静态 HTML。
- `cms.db`：SQLite 数据库文件。

## 前台访问规则

前台页面是静态 HTML：

- `/` 对应 `public/index.html`
- `/post/slug.html` 对应 `public/post/slug.html`
- `/column/slug.html` 对应 `public/column/slug.html`

后台路由仍由 Gin 处理：

- `/adm1n/login`
- `/adm1n/dashboard`
- `/adm1n/posts`
- `/adm1n/columns`
- `/adm1n/tags`
- `/adm1n/settings`

不要把后台页面生成到 `public/`。

## 静态生成机制

静态生成入口在 `handlers/staticgen.go`。

触发场景：

- 程序启动时会调用 `GenerateStatic()`。
- 保存文章后会重新生成静态页。
- 删除文章后会重新生成静态页。
- 切换文章发布状态后会重新生成静态页。
- 定时发布任务到点发布文章后会重新生成静态页。
- 保存栏目后会重新生成静态页。
- 删除栏目后会重新生成静态页。
- 保存网站设置后会重新生成静态页。
- 保存自定义标签后会重新生成静态页。

`GenerateStatic()` 会清理并重建：

- `public/index.html`
- `public/post/`
- `public/column/`

如果直接修改数据库，不会自动触发运行中的程序重新生成静态页。直接改库后建议重启程序，或在后台保存一次网站设置/文章/栏目/自定义标签触发重建。

CLI 写内容类命令会调用缓存失效和静态生成，适合生产运维和后续 AI 自动化开发使用。

## CLI 维护命令

编译后的二进制支持后台和数据库维护 CLI。查看帮助：

```bash
cms help
cms user help
cms post help
cms column help
cms tag help
cms setting help
cms db help
```

常用命令：

```bash
cms serve
cms static
cms migrate
cms db schema
cms db repair
cms db repair --generate-static
```

后台用户：

```bash
cms user list
cms user create --username editor --password secret --nickname 编辑 --admin
cms user password --username admin --password new-secret
cms user delete --id 2
cms reset-admin admin new-secret
```

文章：

```bash
cms post list
cms post get --slug cms-launch
cms post save --title 标题 --slug article-slug --summary 摘要 --content 正文 --column-id 1 --published true
cms post publish --slug article-slug
cms post unpublish --id 1
cms post delete --slug article-slug
```

栏目：

```bash
cms column list
cms column get --slug news
cms column save --name 新闻 --slug news --sort 1
cms column save --id 2 --name 子栏目 --slug child --parent-id 1
cms column delete --slug news
```

自定义标签和设置：

```bash
cms tag list
cms tag set --key service_phone --name 客服电话 --value 400-000-0000
cms setting list
cms setting get --key site_title
cms setting set --key site_title --value 网站标题
```

CLI 注意事项：

- `help`、`version` 和 `<command> help` 不应初始化数据库或写日志。
- CLI 会读取同一份 `config.ini`，生产执行前必须确认工作目录和数据库路径正确。
- `reset-admin` 和 `user password` 只修改已有用户密码，不创建新用户。
- 写内容类命令应在成功后触发 `InvalidateCache()` 和 `GenerateStatic()`。
- 新增 CLI 命令时保持参数风格为 `cms <resource> <action> --key value`，方便脚本和 AI 调用。

## 模板文件

前台模板：

- `templates/index/layout.html`：前台基础布局。
- `templates/index/index.html`：首页、栏目页、文章页共用内容模板。

后台模板：

- `templates/admin/header.html`：后台顶部栏和侧边栏。
- `templates/admin/footer.html`：后台公共底部。
- `templates/admin/login.html`：登录页。
- `templates/admin/dashboard.html`：仪表盘。
- `templates/admin/posts.html`：文章列表。
- `templates/admin/post_edit.html`：文章编辑。
- `templates/admin/columns.html`：栏目列表。
- `templates/admin/column_edit.html`：栏目编辑。
- `templates/admin/settings.html`：网站设置和联系信息。
- `templates/admin/tags.html`：自定义标签列表。
- `templates/admin/tag_edit.html`：自定义标签编辑。

## 模板函数

模板中可用函数在 `main.go` 和 `handlers/staticgen.go` 中注册。

可用函数：

- `asset`：生成 `/static/...` 资源路径，主要用于预留公共静态资源。
- `safe`：输出可信 HTML。
- `md`：把 Markdown 内容渲染为 HTML。
- `render`：解析并渲染字符串中的 Go 模板代码，当前主要用于单页栏目内容。
- `dict`：构造模板参数字典。

前台模板自带资源推荐放在当前模板目录下，并使用绝对路径引用：

```html
<link rel="stylesheet" href="/css/site.css">
<script src="/js/site.js"></script>
<img src="/images/logo.png" alt="Logo">
```

例如当前模板目录是 `templates/index/`，资源放置位置为：

```text
templates/index/css/site.css
templates/index/js/site.js
templates/index/images/logo.png
```

静态生成后会复制为：

```text
public/css/site.css
public/js/site.js
public/images/logo.png
```

不要在前台模板中写 `css/site.css`、`images/logo.png` 这类相对路径。文章页 URL 是 `/post/xxx.html`，相对路径会被浏览器解析成 `/post/css/site.css`，容易导致资源 404。

## 从 demo 制作前台模板

当拿到类似 `demo/` 这样的静态 HTML 模板包时，不要直接把 demo 目录作为前台模板使用。推荐新建一个 `templates/模板名/` 目录，并把静态模板改造成 CMS 可识别的模板。

最小目录结构：

```text
templates/模板名/
templates/模板名/layout.html
templates/模板名/index.html
templates/模板名/css/
templates/模板名/js/
templates/模板名/images/
templates/模板名/fonts/
```

必需文件：

- `layout.html`：前台外层布局，通常放 `<html>`、`<head>`、导航、页脚、公共 CSS 和 JS。
- `index.html`：前台内容入口，首页、栏目页、文章页都通过它渲染。

后台的“前台模板”下拉框只会识别同时包含 `layout.html` 和 `index.html` 的目录。模板目录名不要使用 `admin`，因为 `templates/admin/` 是后台模板目录。

改造步骤：

1. 复制 demo 资源目录到新模板目录。

```text
demo/css      -> templates/模板名/css
demo/js       -> templates/模板名/js
demo/images   -> templates/模板名/images
demo/fonts    -> templates/模板名/fonts
```

2. 将 demo 首页拆成 `layout.html` 和 `index.html`。

`layout.html` 推荐结构：

```html
{{define "front/layout"}}
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "front_title" .}}{{index .settings "site_title"}}{{end}}</title>
    <link rel="stylesheet" href="/css/bootstrap.css">
    <link rel="stylesheet" href="/css/style.css">
</head>
<body>
    {{block "front_content" .}}{{end}}
    <script src="/js/core.min.js"></script>
    <script src="/js/script.js"></script>
</body>
</html>
{{end}}
```

`index.html` 推荐结构：

```html
{{template "front/layout" .}}

{{define "front_content"}}
    {{if .post}}
        <article>
            <h1>{{.post.Title}}</h1>
            {{md .post.Content}}
        </article>
    {{else if .column}}
        {{if .column.IsPage}}
            {{.page_content}}
        {{else}}
            <h1>{{.column.Name}}</h1>
            {{range .posts}}
                <a href="/post/{{.Slug}}.html">{{.Title}}</a>
            {{end}}
        {{end}}
    {{else}}
        <h1>{{index .settings "site_title"}}</h1>
        {{range .posts}}
            <a href="/post/{{.Slug}}.html">{{.Title}}</a>
        {{end}}
    {{end}}
{{end}}
```

3. 把 demo 中写死的菜单改成 CMS 栏目循环。

```html
<nav>
    <a href="/">首页</a>
    {{range .columns}}
        <a href="/column/{{.Slug}}.html">{{.Name}}</a>
    {{end}}
</nav>
```

4. 把 demo 中写死的文章、新闻、产品列表改成 `.posts` 循环。

```html
{{range .posts}}
<article>
    <h2><a href="/post/{{.Slug}}.html">{{.Title}}</a></h2>
    <p>{{.Summary}}</p>
</article>
{{end}}
```

5. 把 demo 中写死的网站名称、电话、邮箱、地址等改成 `.settings` 或 `.custom_tags`。

```html
<strong>{{index .settings "site_title"}}</strong>
<span>{{index .settings "company_phone"}}</span>
<span>{{index .settings "company_email"}}</span>
<span>{{index .custom_tags "service_phone"}}</span>
```

6. 修正所有资源路径为绝对路径。

正确：

```html
<link rel="stylesheet" href="/css/style.css">
<script src="/js/script.js"></script>
<img src="/images/logo.png" alt="Logo">
```

错误：

```html
<link rel="stylesheet" href="css/style.css">
<script src="js/script.js"></script>
<img src="images/logo.png" alt="Logo">
```

7. 清理 demo 演示链接。

不要保留这类原始静态页面链接：

```text
about-us.html
services.html
single-service.html
projects.html
contacts.html
blog-post.html
```

应替换为 CMS 链接：

```html
<a href="/column/{{.Slug}}.html">{{.Name}}</a>
<a href="/post/{{.Slug}}.html">{{.Title}}</a>
<a href="/column/contact.html">联系我们</a>
```

8. 如需把 demo 的内页改成可选单页模板，放在同一模板目录下，文件名使用 `.html`。

示例：

```text
templates/模板名/about-us.html
```

这类文件应作为片段模板使用，不要再写完整的 `{{template "front/layout" .}}`。后台栏目编辑页勾选“单页栏目”后，可以在“单页模板文件”里选择它。

9. 在后台“网站设置”里切换前台模板。

切换后会保存到 `settings.front_template`，并触发 `GenerateStatic()` 重新生成 `public/`。

10. 验证模板。

推荐检查：

```bash
go test ./...
go run .
```

浏览器检查：

```text
/
/column/栏目slug.html
/post/文章slug.html
/css/style.css
/js/script.js
/images/logo.png
```

如果资源 404，优先检查模板里是否还存在相对路径，或资源是否确实放在当前模板目录下。

## 网站设置变量

网站设置保存在 `settings` 表，前台模板通过 `.settings` 读取。

调用方式：

```html
{{index .settings "site_title"}}
```

当前内置变量：

```text
site_title
site_keywords
site_description
site_favicon
site_footer
company_name
company_short_name
company_contact
company_phone
company_email
company_address
company_website
```

示例：

```html
<title>{{index .settings "site_title"}}</title>
<p>电话：{{index .settings "company_phone"}}</p>
<p>邮箱：{{index .settings "company_email"}}</p>
<p>地址：{{index .settings "company_address"}}</p>
```

后台 `网站设置` 页面中，每个字段后面都会显示对应模板变量，方便复制。

## 栏目类型

栏目保存在 `columns` 表。

当前栏目支持两种展示方式：

- 列表栏目：展示当前栏目及其子栏目下的文章列表。
- 单页栏目：展示栏目自身内容，不展示文章列表。

后台入口：

```text
/adm1n/columns
```

编辑栏目时可以勾选 `单页栏目`，并填写 `单页内容 (HTML / 模板代码)`。

单页栏目字段：

```text
is_page
content
```

单页内容会通过 `render` 解析，因此可以直接编写 HTML，也可以调用模板变量和模板函数。

示例：

```html
<section class="about-page">
    <h2>关于我们</h2>
    <p>这里可以直接写 HTML。</p>
    <p>联系电话：{{index .settings "company_phone"}}</p>
    <p>联系邮箱：{{index .settings "company_email"}}</p>
    <p>服务热线：{{index .custom_tags "service_phone"}}</p>
    <img src="{{asset "images/about.jpg"}}" alt="关于我们">
</section>
```

注意事项：

- 单页内容会作为可信 HTML 输出，只允许后台管理员编辑。
- 不要把单页内容编辑权限开放给不可信用户。
- 单页内容中的 `{{...}}` 会被当成模板代码解析。
- 如果只想展示字面量 `{{...}}`，需要避免直接写在会被 `render` 解析的位置。
- 保存栏目后会自动重新生成静态页。

## 自定义标签

自定义标签适合放不属于固定网站设置的变量，例如：

- 首页标语
- 服务热线
- ICP 备案号
- 地图 iframe
- 第三方统计代码
- 页脚额外说明

后台入口：

```text
/adm1n/tags
```

每个自定义标签包含：

- 标签名称：给后台人员看的名称。
- 标签变量：前台模板调用用的 key。
- 标签内容：实际输出的内容。

标签变量规则：

- 只能使用字母、数字、下划线。
- 必须以字母开头。
- 推荐使用小写加下划线，例如 `service_phone`。

后台创建示例：

```text
标签名称：客服电话
标签变量：service_phone
标签内容：400-000-0000
```

前台模板调用：

```html
{{index .custom_tags "service_phone"}}
```

自定义标签会注入到前台模板上下文的 `.custom_tags` map 中。

## 文章发布状态

文章保存在 `posts` 表。

当前文章支持三种发布状态：

- 立即发布：编辑文章时勾选 `发布`。
- 草稿：不勾选 `发布`，也不设置有效的定时发布时间。
- 定时发布：不勾选 `发布`，勾选 `定时发布`，并填写发布时间。

定时发布字段：

```text
scheduled_at
```

程序启动后会调用 `StartPostScheduler()`，后台每分钟检查一次：

```text
published = false
scheduled_at IS NOT NULL
scheduled_at <= 当前时间
```

符合条件的文章会自动改为已发布，清空 `scheduled_at`，并触发 `GenerateStatic()` 重新生成静态页。

注意事项：

- 定时发布依赖 CMS 程序正在运行。
- 如果程序在发布时间点没有运行，下一次启动后会在定时任务检查时补发。
- 勾选 `发布` 时会立即发布，并清空定时发布时间。
- 后台文章列表会区分显示 `已发布`、`草稿`、`定时发布`。

## 修改前台模板

修改前台模板时注意：

- 修改 `templates/index/layout.html` 或 `templates/index/index.html` 后，需要重新生成 `public/` 下的 HTML。
- 最简单方式是重启程序，因为启动时会调用 `GenerateStatic()`。
- 也可以在后台保存一次网站设置、文章、栏目或自定义标签触发静态重建。
- CSS、JS、图片等模板资源放入当前前台模板目录，例如 `templates/index/css/`、`templates/index/js/`、`templates/index/images/`。

前台文章链接推荐：

```html
<a href="/post/{{.Slug}}.html">{{.Title}}</a>
```

前台栏目链接推荐：

```html
<a href="/column/{{.Slug}}.html">{{.Name}}</a>
```

## 修改后台模板

后台公共布局在 `templates/admin/header.html` 和 `templates/admin/footer.html`。

后台样式在：

```text
templates/admin/css/admin.css
```

后台模板引用样式时使用：

```html
/admin-assets/css/admin.css
```

不要把后台页面生成到 `public/`。

## 直接操作数据库

生产环境优先使用后台或 CLI 维护数据，不推荐直接操作 SQLite。只有紧急排查或一次性人工修复时才考虑手工 SQL。

数据库文件：

```text
cms.db
```

SQLite 表名和主要字段：

- `users`：后台用户。
- `columns`：栏目，包含 `is_page` 和 `content` 字段，可用于单页栏目。
- `posts`：文章，包含 `published` 和 `scheduled_at` 字段，可用于立即发布、草稿和定时发布。
- `settings`：网站设置变量。
- `custom_tags`：自定义标签。

直接把栏目设置为单页 SQL 示例：

```sql
UPDATE columns
SET is_page = 1,
    content = '<section><h2>关于我们</h2><p>这里可以写 HTML。</p></section>',
    updated_at = datetime('now')
WHERE slug = 'about';
```

恢复为文章列表栏目：

```sql
UPDATE columns
SET is_page = 0,
    updated_at = datetime('now')
WHERE slug = 'about';
```

直接增加自定义标签 SQL 示例：

直接把文章设置为定时发布 SQL 示例：

```sql
UPDATE posts
SET published = 0,
    scheduled_at = '2026-05-24 10:30:00',
    updated_at = datetime('now')
WHERE slug = 'cms-launch';
```

直接立即发布文章：

```sql
UPDATE posts
SET published = 1,
    scheduled_at = NULL,
    updated_at = datetime('now')
WHERE slug = 'cms-launch';
```

直接增加自定义标签 SQL 示例：

```sql
INSERT INTO custom_tags (name, key, value, created_at, updated_at)
VALUES ('客服电话', 'service_phone', '400-000-0000', datetime('now'), datetime('now'));
```

更新自定义标签：

```sql
UPDATE custom_tags
SET value = '400-111-2222', updated_at = datetime('now')
WHERE key = 'service_phone';
```

删除自定义标签：

```sql
DELETE FROM custom_tags
WHERE key = 'service_phone';
```

直接增加网站设置变量：

```sql
INSERT INTO settings (key, value, created_at, updated_at)
VALUES ('company_phone', '400-000-0000', datetime('now'), datetime('now'));
```

如果 key 已存在，请使用更新：

```sql
UPDATE settings
SET value = '400-000-0000', updated_at = datetime('now')
WHERE key = 'company_phone';
```

注意事项：

- `settings.key` 和 `custom_tags.key` 都有唯一约束。
- 直接改库不会自动刷新运行中的缓存。
- 直接改库不会自动重新生成 `public/` 下静态 HTML。
- 直接改库后建议重启程序，或在后台保存一次相关配置触发 `GenerateStatic()`。
- 如果必须直接改库，改完后优先执行 `cms static` 或重启服务。

## 生产兼容和数据库升级规则

本项目已经用于生产环境。后续维护时必须遵守：

- 不带参数的 `cms` 必须继续启动 Web 服务，不能改变生产启动方式。
- 后台 `/adm1n/...` 路由、前台 `/post/slug.html` 和 `/column/slug.html` URL 规则不应随意改变。
- 数据库结构修改必须写进 Go 代码，通过 `database.Init()`、`AutoMigrate` 或显式迁移函数自动完成。
- 不能只给手工 SQL 作为数据库升级方案；编译后的二进制必须能兼容或升级已有生产 `cms.db`。
- 迁移逻辑必须幂等，重复启动或重复执行不应破坏数据。
- 新增字段优先允许空值或提供安全默认值，兼容历史数据。
- 不随意删除字段、重命名字段、改变字段含义、改变唯一约束或破坏旧模板变量。
- 需要修复旧数据时，优先新增 Go 迁移或 `cms db repair` 子命令。
- 升级前后建议执行 `cms db schema` 记录并核对数据库结构。
- 修改 session、登录、密码、安全策略时要评估生产影响，避免无提示让所有管理员登录态失效。
- 修改静态生成逻辑时要避免误删生产手工资源；`public/` 是生成目录，不要存放需要长期保留的人工文件。

## 开发注意事项

- 修改 Go 文件后运行 `gofmt -w .`。
- 修改 Go 代码后运行 `go test ./...`。
- 修改模板后建议短暂运行 `go run .`，确认模板能正常解析。
- 不要把用户上传或手工维护的资源放进 `public/post` 或 `public/column`，这些目录会被静态生成清理。
- 前台模板资源应放在对应前台模板目录，后台样式资源应放在 `templates/admin/css/`。
- 前台模板中如需输出可信 HTML，可使用 `safe`，但不要对不可信用户输入滥用。
- 文章内容 Markdown 会通过 `md` 渲染为 HTML。
- 单页栏目内容会通过 `render` 渲染，可写 HTML 和模板变量，但必须视为可信内容。

## 常见排查

前台页面没有更新：

- 检查是否重新生成了 `public/`。
- 重启程序或在后台保存一次网站设置触发重建。
- 检查 `config.ini` 中 `[static] enable = true`。

资源 404：

- 前台模板资源确认在当前前台模板目录下，例如 `templates/index/css/`、`templates/index/js/`、`templates/index/images/`。
- 后台 CSS 确认在 `templates/admin/css/`，并通过 `/admin-assets/css/...` 引用。
- 不要使用相对路径。

自定义标签不显示：

- 检查后台 `/adm1n/tags` 中是否已创建标签。
- 检查模板变量是否写成 `{{index .custom_tags "变量名"}}`。
- 保存标签后确认静态页已重新生成。

端口占用：

- 默认端口是 `8080`。
- 如果启动提示端口已占用，修改 `config.ini` 中 `[server] port`，或停止已有进程。
