# GoCMS CLI 使用说明

编译后的二进制支持 CLI。为兼容生产环境，不带参数时等同于启动 Web 服务。

## 基础命令

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
- `cms refresh`：向正在运行的 `cms serve` 发送刷新请求，由服务进程热更新 `public/` 静态文件。
- `cms static` / `cms generate-static`：重新生成前台静态页面。
- `cms migrate`：初始化数据库并执行自动迁移和默认数据补齐。
- `cms help`：显示帮助。
- `cms version`：显示版本号。

## 用户维护

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

## 文章维护

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

## 栏目维护

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

## 自定义标签

```bash
cms tag help
cms tag list
cms tag set --key service_phone --name 客服电话 --value 400-000-0000
cms tag delete --key service_phone
```

## 网站设置

```bash
cms setting help
cms setting list
cms setting get --key site_title
cms setting set --key site_title --value 网站标题
```

## 数据库维护

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

## 生产注意事项

- CLI 会读取同一份 `config.ini`，执行前确认工作目录和数据库路径正确。
- 写内容类命令成功后会触发缓存失效和静态生成。
- `help`、`version` 和 `<command> help` 不会初始化数据库或写日志。
- 数据库结构升级必须随 Go 代码发布，不要只依赖手工 SQL。
- 直接修改 SQLite 后，需要重启服务或执行 `cms static`。
