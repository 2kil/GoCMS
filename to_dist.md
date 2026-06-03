# GoCMS 部署说明

GoCMS 是一个基于 SQLite 的轻量 CMS。后台动态管理内容，前台生成静态 HTML 到 `public/` 目录。

## 文件说明

打包目录通常包含：

```text
cms              Linux 可执行文件
cms.exe          Windows 可执行文件
config.ini       配置文件
templates/       后台模板和前台模板
static/          /static 公共静态资源
README.md        本说明文件
```

程序首次运行会自动创建 `cms.db`、`cms.log` 和 `public/`。

## 启动服务

Linux：

```bash
./cms
```

Windows：

```powershell
.\cms.exe
```

不带参数时等同于：

```bash
cms serve
```

默认访问地址：

```text
http://localhost:8080/
http://localhost:8080/adm1n/login
```

默认后台账号：

```text
admin / G0u8NmtXSsFmDwxDCl
```

生产环境上线后请及时修改后台密码和 `[session] secret`。

## 配置文件

程序优先读取当前工作目录下的 `config.ini`。如果不存在，会读取可执行文件同目录下的 `config.ini`。

常用配置：

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

相对路径会按 `config.ini` 所在目录解析。

## 常用 CLI

```bash
cms help
cms version
cms serve
cms refresh
cms static
cms migrate
```

说明：

- `cms` / `cms serve`：启动 Web 服务。
- `cms refresh`：向正在运行的 `cms serve` 发送刷新请求，由服务进程热更新 `public/` 静态文件。
- `cms static`：重新生成前台静态页面。
- `cms migrate`：初始化数据库并执行自动迁移和默认数据补齐。
- `cms help`：显示帮助。
- `cms version`：显示版本号。

## 用户维护

```bash
cms user list
cms user create --username editor --password secret --nickname 编辑 --admin
cms user password --username admin --password new-secret
cms user delete --id 2
cms reset-admin admin new-secret
```

## 内容维护

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
cms column save --id 2 --name 顶级栏目 --slug top --no-parent
cms column delete --slug news
```

自定义标签和网站设置：

```bash
cms tag list
cms tag set --key service_phone --name 客服电话 --value 400-000-0000
cms tag delete --key service_phone
cms setting list
cms setting get --key site_title
cms setting set --key site_title --value 网站标题
```

数据库维护：

```bash
cms db schema
cms db repair
cms db repair --generate-static
```

## 前台模板

前台模板位于 `templates/` 下。一个有效模板目录必须同时包含：

```text
layout.html
index.html
```

例如：

```text
templates/company/layout.html
templates/company/index.html
templates/company/css/style.css
templates/company/js/script.js
templates/company/images/logo.png
```

前台资源请使用绝对路径：

```html
<link rel="stylesheet" href="/css/style.css">
<script src="/js/script.js"></script>
<img src="/images/logo.png" alt="Logo">
```

不要使用 `css/style.css`、`js/script.js`、`images/logo.png` 这类相对路径，否则文章页和栏目页容易出现资源 404。

后台“网站设置”中可以切换前台模板。保存后会重新生成静态页面。

## 从 demo 改造模板

不要直接把 `demo/` 目录作为前台模板使用。推荐新建 `templates/模板名/`，再复制资源并拆分 HTML。

基本步骤：

1. 新建 `templates/模板名/`。
2. 复制 `demo/css`、`demo/js`、`demo/images`、`demo/fonts` 到新模板目录。
3. 把 demo 首页公共结构放入 `layout.html`。
4. 把首页、栏目页、文章页内容入口放入 `index.html`。
5. 把写死导航替换为 `.columns` 循环。
6. 把写死新闻、博客、产品列表替换为 `.posts` 循环。
7. 把网站名称、电话、邮箱、地址替换为 `.settings` 或 `.custom_tags`。
8. 把资源路径改为 `/css/...`、`/js/...`、`/images/...`。
9. 清理 `about-us.html`、`services.html`、`contacts.html` 等 demo 静态链接。
10. 在后台“网站设置”切换新模板并保存。

栏目导航示例：

```html
<a href="/">首页</a>
{{range .columns}}
    <a href="/column/{{.Slug}}.html">{{.Name}}</a>
{{end}}
```

文章列表示例：

```html
{{range .posts}}
<article>
    <h2><a href="/post/{{.Slug}}.html">{{.Title}}</a></h2>
    <p>{{.Summary}}</p>
</article>
{{end}}
```

网站设置示例：

```html
{{index .settings "site_title"}}
{{index .settings "company_phone"}}
{{index .settings "company_email"}}
{{index .custom_tags "service_phone"}}
```

## 静态页面

前台静态页面生成到：

```text
public/index.html
public/post/*.html
public/column/*.html
```

不要把需要长期保留的人工文件放进 `public/post/` 或 `public/column/`。

如果直接修改数据库，不会自动刷新运行中的缓存，也不会自动重新生成静态页。直接改库后请重启服务，或执行：

```bash
cms static
```

如果服务正在运行，推荐执行：

```bash
cms refresh
```

## 升级注意事项

- 备份旧的 `cms.db`、`config.ini` 和 `templates/`。
- 替换可执行文件前确认系统平台使用正确的二进制。
- 不带参数的 `cms` 会继续启动 Web 服务。
- 升级前后可执行 `cms db schema` 核对数据库结构。
- 如果修改了模板，保存网站设置或执行 `cms static` 重新生成前台页面。
