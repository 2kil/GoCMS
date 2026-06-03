# GoCMS

GoCMS 是一个基于 Go、Gin、GORM 和 SQLite 的轻量 CMS。后台使用 Gin 动态渲染管理页面，前台以模板生成静态 HTML，并由 `public/` 目录对外提供访问。

## 功能概览

- 后台登录、退出和 Session 鉴权。
- 文章管理：创建、编辑、删除、发布/取消发布、封面图、摘要、栏目归属、Markdown 内容。
- 文章发布方式：草稿、立即发布、定时发布。
- 栏目管理：树形父子栏目、排序、列表栏目、单页栏目。
- 单页栏目支持直接填写 HTML/模板代码，也支持选择当前前台模板目录下的单页模板片段。
- 网站设置：站点 SEO、页脚、公司信息、前台模板选择、当前账号用户名和密码修改。
- 自定义标签：通过模板变量输出任意后台维护内容。
- 前台模板：支持多个模板目录，后台可切换。
- 静态生成：生成首页、文章页、栏目页，并复制当前前台模板的静态资源。

## 技术栈

- Go
- Gin
- gin-contrib/sessions
- GORM
- glebarez/sqlite
- gomarkdown/markdown
- bcrypt

## 快速开始

```bash
go run .
```

不带参数时等同于启动 Web 服务，也可以显式执行：

```bash
go run . serve
```

启动后访问：

```text
http://localhost:8080/
http://localhost:8080/adm1n/login
```

默认后台账号：

```text
admin / G0u8NmtXSsFmDwxDCl
```

第一次启动会自动创建 SQLite 数据库、迁移数据表、创建默认管理员和默认网站设置。

## CLI 使用说明

编译后的二进制支持 CLI。为兼容已有生产启动方式，不带参数时等同于启动 Web 服务。

### 基础命令

```bash
cms
cms serve
cms help
cms version
cms refresh
cms static
cms generate-static
cms migrate
```

说明：

- `cms` / `cms serve`：启动 Web 服务。
- `cms refresh`：向正在运行的 `cms serve` 发送刷新请求，由服务进程热更新 `public/` 静态文件，避免 Windows 下另一个进程替换目录导致文件占用。
- `cms static` / `cms generate-static`：重新生成前台静态页面。
- `cms migrate`：初始化数据库并执行自动迁移和默认数据补齐。
- `cms help`：显示帮助。
- `cms version`：显示版本号。

### 用户维护

```bash
cms user help
cms user list
cms user create --username editor --password secret --nickname 编辑 --admin
cms user password --username admin --password new-secret
cms user delete --id 2
cms reset-admin admin new-secret
```

说明：

- `reset-admin` 和 `user password` 只修改已有用户密码，不创建新用户。
- `--admin` 是布尔参数，传入即表示管理员。

### 文章维护

```bash
cms post help
cms post list
cms post get --slug cms-launch
cms post get --id 1
cms post save --title 标题 --slug article-slug --summary 摘要 --content 正文 --column-id 1 --published true
cms post save --id 1 --title 新标题 --slug article-slug --content 新正文
cms post publish --slug article-slug
cms post unpublish --id 1
cms post delete --slug article-slug
```

常用参数：

- `--id`：文章 ID。
- `--slug`：文章 URL 标识。
- `--title`：标题。
- `--summary`：摘要。
- `--content`：正文。
- `--cover`：封面图地址。
- `--column-id`：栏目 ID。
- `--no-column`：清空栏目。
- `--published true|false`：发布状态。
- `--scheduled-at "2026-05-24 10:30"`：定时发布时间。

### 栏目维护

```bash
cms column help
cms column list
cms column get --slug news
cms column get --id 1
cms column save --name 新闻 --slug news --sort 1
cms column save --id 2 --name 子栏目 --slug child --parent-id 1
cms column save --id 2 --name 顶级栏目 --slug top --no-parent
cms column delete --slug news
```

常用参数：

- `--id`：栏目 ID。
- `--slug`：栏目 URL 标识。
- `--name`：栏目名称。
- `--parent-id`：父栏目 ID。
- `--no-parent`：设为顶级栏目。
- `--sort`：排序值。
- `--is-page true|false`：是否单页栏目。
- `--page-template`：单页模板文件。
- `--content`：单页内容。

栏目保存会校验父栏目不能选择自身或自己的子栏目。

### 自定义标签

```bash
cms tag help
cms tag list
cms tag set --key service_phone --name 客服电话 --value 400-000-0000
cms tag delete --key service_phone
```

### 网站设置

```bash
cms setting help
cms setting list
cms setting get --key site_title
cms setting set --key site_title --value 网站标题
```

### 数据库维护

```bash
cms db help
cms db init
cms db migrate
cms db schema
cms db repair
cms db repair --generate-static
```

说明：

- `db init` / `db migrate`：初始化数据库并执行自动迁移。
- `db schema`：输出当前 SQLite 数据库结构。
- `db repair`：执行内置安全修复，例如补齐空 slug、空标题、空昵称、栏目排序和自定义标签名称。
- `db repair --generate-static`：修复后重新生成静态页面。

### CLI 生产注意事项

- CLI 会读取同一份 `config.ini`，执行前确认工作目录和数据库路径正确。
- 写内容类命令成功后会触发缓存失效和静态生成。
- `help`、`version` 和 `<command> help` 不会初始化数据库或写日志。
- 数据库结构升级必须随 Go 代码发布，不要只依赖手工 SQL。
- 直接修改 SQLite 后，需要重启服务或执行 `cms static`。

## 常用命令

```bash
go test ./...
```

```bash
gofmt -w .
```

```bash
./build.ps1
```

## 配置文件

项目会优先读取当前工作目录下的 `config.ini`；如果不存在，会尝试读取可执行文件同目录下的 `config.ini`；仍不存在时使用代码中的默认配置。

当前仓库的 `config.ini`：

```ini
[server]
port = 8080

[database]
path = cms.db

[session]
secret = cms-secret-key-2026

[log]
file = cms.log

[static]
enable = true
dir = public
```

代码默认值略有不同：如果没有配置文件，`static.enable` 默认为 `false`，不会生成和读取前台静态页。

## 目录结构

```text
config/              配置加载和模板路径解析
database/            数据库初始化、自动迁移、默认数据写入
handlers/            后台、前台、静态生成、模板函数、定时发布
middleware/          后台登录鉴权
models/              GORM 数据模型
templates/admin/     后台模板和后台 CSS
templates/index/     默认前台模板
static/              /static 公共静态资源
public/              生成后的前台静态文件
main.go              应用入口和路由注册
cli.go               CLI 子命令实现
config.ini           默认配置
cms.db               默认 SQLite 数据库文件
```

## 路由说明

后台路由固定使用 `/adm1n` 前缀：

```text
GET  /adm1n/login
POST /adm1n/login
GET  /adm1n/logout
GET  /adm1n/dashboard
GET  /adm1n/posts
GET  /adm1n/posts/edit/:id
POST /adm1n/posts/save/:id
GET  /adm1n/posts/delete/:id
GET  /adm1n/posts/toggle/:id
GET  /adm1n/columns
GET  /adm1n/columns/edit/:id
POST /adm1n/columns/save/:id
POST /adm1n/columns/reorder
GET  /adm1n/columns/delete/:id
GET  /adm1n/tags
GET  /adm1n/tags/edit/:id
POST /adm1n/tags/save/:id
GET  /adm1n/tags/delete/:id
GET  /adm1n/settings
POST /adm1n/settings
POST /adm1n/settings/account
```

未匹配后台路由时，程序会从 `public/` 读取静态文件：

```text
/                  -> public/index.html
/post/slug.html    -> public/post/slug.html
/column/slug.html  -> public/column/slug.html
/css/site.css      -> public/css/site.css
```

## 静态生成

静态生成由 `handlers.GenerateStatic()` 完成，并受 `[static] enable` 控制。

会生成：

```text
public/index.html
public/post/*.html
public/column/*.html
```

生成前会清理 `public/` 下所有现有内容，然后复制当前前台模板目录中的非 `.html` 文件到 `public/`。因此不要把需要长期保留的手工文件放在 `public/` 下。

触发静态全量生成的场景：

- 程序启动。
- 保存、删除或切换文章发布状态。
- 定时发布任务发布文章。
- 保存、删除或排序栏目。
- 保存网站设置。
- 保存或删除自定义标签。

如果直接修改数据库，不会自动通知正在运行的程序刷新缓存和重新生成静态页。直接改库后建议重启程序，或在后台保存一次相关内容触发生成。

## 文章管理

文章字段包括：

```text
title
slug
summary
content
cover_image
published
scheduled_at
column_id
user_id
```

说明：

- `slug` 为空时会自动生成 UUID。
- 文章内容通过模板函数 `md` 渲染为 HTML。
- 勾选发布时立即发布，并清空定时发布时间。
- 不勾选发布、但勾选定时发布并填写时间时，会保存为定时发布。
- 定时发布任务启动后立即检查一次，之后每分钟检查一次。
- 已发布文章才会生成前台文章页和出现在前台列表中。

## 栏目管理

栏目字段包括：

```text
name
slug
is_page
page_template
content
sort_order
parent_id
```

说明：

- 栏目支持父子层级。
- 列表栏目会展示当前栏目及其子栏目下的已发布文章。
- 单页栏目会展示栏目自身内容，不展示文章列表。
- 单页栏目可以选择当前前台模板目录下除 `index.html`、`layout.html` 以外的 `.html` 文件作为片段模板。
- 未选择单页模板片段时，会渲染栏目 `content` 字段。

## 网站设置

内置网站设置 key：

```text
site_title
site_keywords
site_description
site_favicon
site_footer
front_template
company_name
company_short_name
company_contact
company_phone
company_email
company_address
company_website
```

模板中通过 `.settings` 读取：

```html
{{index .settings "site_title"}}
{{index .settings "company_phone"}}
```

后台设置页也提供当前账号用户名和密码修改。修改用户名或密码时需要验证旧密码。

## 自定义标签

自定义标签适合维护非固定字段，例如客服电话、ICP备案号、统计代码、地图 iframe 等。

标签变量规则：

```text
^[a-zA-Z][a-zA-Z0-9_]*$
```

模板中通过 `.custom_tags` 读取：

```html
{{index .custom_tags "service_phone"}}
```

## 前台模板

前台模板目录位于 `templates/` 下，且必须同时包含 `layout.html` 和 `index.html`。

`templates/admin/` 是后台模板目录，不会作为前台模板。默认前台模板是 `templates/index/`。后台“网站设置”中的前台模板下拉框来自所有有效前台模板目录。

### 从 demo 改造 HTML 模板

不要直接把 `demo/` 目录作为前台模板使用。推荐新建一个模板目录，例如：

```text
templates/模板名/
templates/模板名/layout.html
templates/模板名/index.html
templates/模板名/css/
templates/模板名/js/
templates/模板名/images/
templates/模板名/fonts/
```

模板目录名不要使用 `admin`。

### 1. 新建模板目录

在 `templates/` 下新建目录，例如：

```text
templates/company/
```

最终目录至少应包含：

```text
templates/company/layout.html
templates/company/index.html
```

### 2. 复制静态资源

把 `demo/` 里的资源目录复制到新的模板目录中：

```text
demo/css      -> templates/company/css
demo/js       -> templates/company/js
demo/images   -> templates/company/images
demo/fonts    -> templates/company/fonts
```

如果模板包还有 `assets/`、`vendor/` 等资源目录，也可以复制到模板目录下。

### 3. 拆分 HTML

把 demo 首页拆成两个文件：

- `layout.html`：放公共 HTML 结构，例如 `<!DOCTYPE html>`、`<head>`、导航、页脚、公共 CSS 和 JS。
- `index.html`：放首页、栏目页、文章页的主要内容入口。

`layout.html` 示例：

```html
{{define "front/layout"}}
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "front_title" .}}{{index .settings "site_title"}}{{end}}</title>
    <meta name="keywords" content="{{index .settings "site_keywords"}}">
    <meta name="description" content="{{index .settings "site_description"}}">
    <link rel="stylesheet" href="/css/style.css">
</head>
<body>
    <header>
        <a href="/">{{index .settings "site_title"}}</a>
        <nav>
            <a href="/">首页</a>
            {{range .columns}}
                <a href="/column/{{.Slug}}.html">{{.Name}}</a>
            {{end}}
        </nav>
    </header>

    {{block "front_content" .}}{{end}}

    <footer>
        {{index .settings "site_footer"}}
    </footer>
    <script src="/js/script.js"></script>
</body>
</html>
{{end}}
```

`index.html` 示例：

```html
{{template "front/layout" .}}

{{define "front_content"}}
    {{if .post}}
        <article>
            <h1>{{.post.Title}}</h1>
            <div>{{md .post.Content}}</div>
        </article>
    {{else if .column}}
        {{if .column.IsPage}}
            {{.page_content}}
        {{else}}
            <section>
                <h1>{{.column.Name}}</h1>
                {{range .posts}}
                    <article>
                        <h2><a href="/post/{{.Slug}}.html">{{.Title}}</a></h2>
                        <p>{{.Summary}}</p>
                    </article>
                {{end}}
            </section>
        {{end}}
    {{else}}
        <section>
            <h1>{{index .settings "site_title"}}</h1>
            {{range .posts}}
                <article>
                    <h2><a href="/post/{{.Slug}}.html">{{.Title}}</a></h2>
                    <p>{{.Summary}}</p>
                </article>
            {{end}}
        </section>
    {{end}}
{{end}}
```

### 4. 替换写死导航

把 demo 中写死的菜单改成 CMS 栏目循环。

原始静态链接通常类似：

```html
<a href="about-us.html">关于我们</a>
<a href="services.html">服务项目</a>
<a href="contacts.html">联系我们</a>
```

应改成：

```html
<a href="/">首页</a>
{{range .columns}}
    <a href="/column/{{.Slug}}.html">{{.Name}}</a>
{{end}}
```

如果某个链接需要固定指向某个栏目，可以使用该栏目的 slug：

```html
<a href="/column/contact.html">联系我们</a>
```

### 5. 替换文章和新闻列表

把 demo 中写死的新闻、博客、产品或案例列表改成 `.posts` 循环。

```html
{{range .posts}}
<article>
    <h2><a href="/post/{{.Slug}}.html">{{.Title}}</a></h2>
    <p>{{.Summary}}</p>
</article>
{{end}}
```

文章详情页通过 `.post` 判断：

```html
{{if .post}}
<article>
    <h1>{{.post.Title}}</h1>
    <div>{{md .post.Content}}</div>
</article>
{{end}}
```

### 6. 替换网站信息变量

把 demo 中写死的网站名称、电话、邮箱、地址等替换为 `.settings` 或 `.custom_tags`。

常用网站设置变量：

```html
{{index .settings "site_title"}}
{{index .settings "site_keywords"}}
{{index .settings "site_description"}}
{{index .settings "site_footer"}}
{{index .settings "company_name"}}
{{index .settings "company_short_name"}}
{{index .settings "company_contact"}}
{{index .settings "company_phone"}}
{{index .settings "company_email"}}
{{index .settings "company_address"}}
{{index .settings "company_website"}}
```

自定义标签示例：

```html
{{index .custom_tags "service_phone"}}
```

### 7. 修正资源路径

前台模板资源必须使用绝对路径。

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

原因：文章页 URL 是 `/post/xxx.html`，相对路径 `css/style.css` 会被浏览器解析成 `/post/css/style.css`，容易导致资源 404。

### 8. 清理 demo 演示链接

不要保留原始静态页面链接，例如：

```text
about-us.html
services.html
single-service.html
projects.html
contacts.html
blog-post.html
```

应替换为 CMS 路径：

```html
<a href="/column/{{.Slug}}.html">{{.Name}}</a>
<a href="/post/{{.Slug}}.html">{{.Title}}</a>
<a href="/column/contact.html">联系我们</a>
```

### 9. 制作可选单页模板

如果 demo 中有“关于我们”“联系我们”等内页，可以把它们改造成可选单页模板。

示例：

```text
templates/company/about.html
templates/company/contact.html
```

这类文件应作为片段模板使用，不要再写完整的 `{{template "front/layout" .}}`，也不要包含完整的 `<!DOCTYPE html>`、`<html>`、`<head>`、`<body>`。

单页模板示例：

```html
<section class="about-page">
    <h1>关于我们</h1>
    <p>{{index .settings "company_name"}}</p>
    <p>电话：{{index .settings "company_phone"}}</p>
    <p>邮箱：{{index .settings "company_email"}}</p>
</section>
```

后台栏目编辑页中勾选“单页栏目”后，可以在“单页模板文件”里选择这些文件。

### 10. 切换前台模板

进入后台“网站设置”，在“前台模板”下拉框中选择新模板目录。

保存后会写入：

```text
settings.front_template
```

保存网站设置会触发静态页面重新生成。

### 模板结构参考

推荐结构：

```text
templates/模板名/layout.html
templates/模板名/index.html
templates/模板名/about.html
templates/模板名/css/site.css
templates/模板名/js/site.js
templates/模板名/images/logo.png
```

前台模板资源建议使用绝对路径：

```html
<link rel="stylesheet" href="/css/site.css">
<script src="/js/site.js"></script>
<img src="/images/logo.png" alt="Logo">
```

静态生成时会把当前前台模板目录里的非 `.html` 文件复制到 `public/` 对应路径。

### 模板验证

改造完成后建议执行：

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

如果资源 404，优先检查模板中是否还存在相对路径，资源是否确实放在当前前台模板目录下，后台“网站设置”中是否已经切换到新模板，以及是否已经重新生成 `public/`。

## 模板变量和函数

前台模板常用变量：

```text
.posts        文章列表
.post         当前文章页文章
.columns      栏目列表
.column       当前栏目页栏目
.settings     网站设置 map
.custom_tags  自定义标签 map
.page_content 单页栏目渲染后的 HTML
```

可用模板函数：

```text
safe    输出可信 HTML
asset   生成 /static/... 资源路径；已是 /、http://、https:// 开头时原样返回
dict    构造 map
list    构造列表
mod     取模，除数为 0 时返回 0
md      将 Markdown 渲染为 HTML
render  渲染字符串模板
```

注意：`render` 用于栏目单页内容时，只注册了 `safe` 和 `asset` 两个函数；完整前台模板和单页模板文件可使用上面的全部函数。

## 数据库维护

生产环境优先使用后台或 CLI 维护数据，不推荐直接修改 SQLite。常用 CLI：

```bash
cms db schema
cms db repair
cms setting set --key site_title --value 网站标题
cms tag set --key service_phone --name 客服电话 --value 400-000-0000
cms post publish --slug cms-launch
cms static
```

数据库文件：

```text
cms.db
```

SQLite 表名和主要用途：

- `users`：后台用户。
- `columns`：栏目，包含父子关系、单页配置和排序。
- `posts`：文章，包含发布状态、定时发布时间、栏目和作者。
- `settings`：网站设置变量。
- `custom_tags`：自定义标签。

注意事项：

- `settings.key` 和 `custom_tags.key` 都有唯一约束。
- CLI 会读取当前工作目录或可执行文件目录下的 `config.ini`，生产执行前先确认数据库路径正确。
- 直接改库不会自动刷新运行中的缓存，也不会自动重新生成 `public/` 下静态 HTML。
- 如果必须直接改库，改完后需要重启程序，或执行 `cms static` 重新生成静态页。
- 涉及数据库结构变更时，必须把迁移逻辑写进 Go 代码并随程序发布，不能只依赖手工 SQL。

## 生产升级规则

本项目已用于生产环境，后续修改必须考虑现有编译产物、配置文件、模板和 SQLite 数据库的兼容性。

- 不带参数的 `cms` 必须继续启动 Web 服务，不能改变生产启动方式。
- 后台路由 `/adm1n/...`、前台 URL `/post/slug.html` 和 `/column/slug.html` 不应随意改变。
- 数据库字段只能做向前兼容升级；不要随意删除字段、改字段含义或破坏已有唯一约束。
- 新增字段应能通过 `database.Init()` 中的 `AutoMigrate` 或显式 Go 迁移自动升级旧库。
- 需要修复历史数据时，应添加幂等 Go 代码或 CLI 修复命令，例如 `cms db repair`，不要只给手工 SQL。
- 修改模板变量时必须保留旧变量，除非确认没有生产模板依赖。
- 修改静态生成逻辑时要避免误删生产手工资源；`public/` 是生成目录，不应存放需要长期保留的人工文件。
- 修改 session、密码、登录、安全策略时要评估生产影响，避免无提示踢掉全部登录态。
- 升级前建议执行 `cms db schema` 记录旧库结构，升级后再次执行核对。

## 开发和维护注意事项

- 修改 Go 代码后运行 `gofmt -w .` 和 `go test ./...`。
- 修改前台模板后需要重新生成 `public/`，最简单方式是重启程序或在后台保存一次相关内容。
- 后台 CSS 通过 `/admin-assets/css/...` 访问，文件位于 `templates/admin/css/`。
- `/static/...` 映射到仓库根目录下的 `static/`。
- `public/` 是生成目录，可能被静态生成清空。
- `cms.log` 为默认运行日志文件。
- 生产环境应修改默认后台密码和 `session.secret`。

## 打包

仓库包含 `build.ps1`，可用于 Windows 环境打包。打包后需要确保可执行文件同目录或工作目录中包含运行所需的 `config.ini`、`templates/`、`static/` 等文件。
