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
admin / admin123
```

第一次启动会自动创建 SQLite 数据库、迁移数据表、创建默认管理员和默认网站设置。

## CLI 命令

编译后的二进制支持 CLI。为兼容已有生产启动方式，不带命令仍会启动 Web 服务。

```bash
cms serve
```

```bash
cms static
```

```bash
cms generate-static
```

```bash
cms migrate
```

```bash
cms reset-admin <username> <password>
```

```bash
cms user list
cms user create --username editor --password secret --nickname 编辑 --admin
cms user password --username admin --password new-secret
cms user delete --id 2
```

```bash
cms post list
cms post get --slug cms-launch
cms post save --title 标题 --slug article-slug --summary 摘要 --content 正文 --column-id 1 --published true
cms post publish --slug article-slug
cms post unpublish --id 1
cms post delete --slug article-slug
```

```bash
cms column list
cms column get --slug news
cms column save --name 新闻 --slug news --sort 1
cms column save --id 2 --name 子栏目 --slug child --parent-id 1
cms column delete --slug news
```

```bash
cms tag list
cms tag set --key service_phone --name 客服电话 --value 400-000-0000
cms tag delete --key service_phone
```

```bash
cms setting list
cms setting get --key site_title
cms setting set --key site_title --value 网站标题
```

```bash
cms db init
cms db migrate
cms db schema
cms db repair
cms db repair --generate-static
```

```bash
cms version
```

```bash
cms help
```

说明：

- `serve`：启动 Web 服务，是默认命令。
- `static` / `generate-static`：初始化配置和数据库后重新生成前台静态页面。
- `migrate`：初始化配置和数据库，执行 GORM 自动迁移和内置默认数据补齐。
- `reset-admin`：重置已有后台用户密码，不会创建新用户。
- `user`：维护后台用户，支持列表、创建、改密码、删除。
- `post`：维护文章，支持列表、查看、保存、删除、发布、取消发布。
- `column`：维护栏目，支持列表、查看、保存、删除，并校验父栏目不能指向自身或子栏目。
- `tag`：维护自定义标签，支持列表、设置、删除。
- `setting`：维护网站设置，支持列表、查看、设置。
- `db init` / `db migrate`：初始化配置和数据库，执行自动迁移和默认数据补齐。
- `db schema`：输出当前 SQLite 数据库表结构，便于升级前后核对。
- `db repair`：执行内置安全修复，例如补齐空 slug、空标题、空昵称和栏目排序；加 `--generate-static` 会修复后重新生成静态页。
- `version`：输出编译版本，默认是 `dev`。
- CLI 命令会读取同一份 `config.ini`，并使用配置中的数据库路径、静态目录和模板目录。
- 数据库结构升级必须随 Go 代码发布；生产环境不要依赖手工 SQL 作为唯一升级方式。

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

前台模板目录位于 `templates/` 下，且必须同时包含：

```text
layout.html
index.html
```

`templates/admin/` 是后台模板目录，不会作为前台模板。

默认前台模板是 `templates/index/`。后台“网站设置”中的前台模板下拉框来自所有有效前台模板目录。

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
