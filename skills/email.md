# Email CLI Client Skill

Email CLI 客户端，用于管理邮箱账户、同步邮件、发送邮件等操作。

## 触发条件

当用户询问邮件相关操作时使用：
- 邮件同步、发送、读取
- 邮箱账户管理
- 邮件搜索、文件夹管理
- 关键词：email、邮件、邮箱、imap、smtp、pop3

## 常用命令

### 账户管理

```bash
# 列出所有账户
email account list

# 添加账户
email account add --name "账户名称" --email "your@email.com" \
  --imap-host "imap.example.com" --imap-port 993 \
  --smtp-host "smtp.example.com" --smtp-port 465 \
  --username "your@email.com" --password "your-password"

# 删除账户
email account delete --id <account-id>

# 显示账户详情
email account show --id <account-id>
```

### 邮件同步

```bash
# 同步所有账户邮件
email sync

# 同步指定账户
email sync --account <account-id>

# 同步指定文件夹
email sync --account <account-id> --folder "INBOX"
```

### 邮件列表与读取

```bash
# 列出邮件
email list --account <account-id> --folder "INBOX"

# 读取邮件
email read --id <message-id>

# JSON 输出（AI 友好）
email list --account <account-id> --output json
email read --id <message-id> --output json
```

### 邮件发送

```bash
# 发送简单邮件
email send --to "recipient@example.com" --subject "主题" --body "正文"

# 发送带附件
email send --to "recipient@example.com" --subject "主题" --body "正文" \
  --attachment "/path/to/file.pdf"

# 从文件读取正文
email send --to "recipient@example.com" --subject "主题" \
  --body-file "/path/to/body.txt"
```

### 搜索

```bash
# 搜索邮件
email search "关键词"

# 高级搜索
email search --from "sender@example.com" --subject "重要" --since "2024-01-01"

# 在指定账户搜索
email search "关键词" --account <account-id>
```

### 文件夹管理

```bash
# 列出文件夹
email folders --account <account-id>

# 创建文件夹
email folders create --account <account-id> --name "新文件夹"
```

## 输出格式

支持两种输出格式：
- 默认：人类可读的表格格式
- JSON：`--output json` 适合 AI 解析

## 示例用法

```bash
# AI 助手常用操作流程
# 1. 列出账户
email account list --output json

# 2. 同步最新邮件
email sync --account <id>

# 3. 获取收件箱列表
email list --account <id> --folder "INBOX" --output json

# 4. 读取特定邮件
email read --id <msg-id> --output json

# 5. 回复邮件
email send --to "sender@example.com" --subject "Re: 原主题" --body "回复内容"
```

## 数据存储

- 数据库位置：`~/.email/email.db`
- 支持 FTS5 全文搜索
- SQLite 格式，便于备份和迁移

## 注意事项

1. 密码以明文存储在数据库中，请确保系统安全
2. 首次使用需要先添加邮箱账户
3. 同步大量邮件可能需要较长时间
4. 发送邮件前确保 SMTP 配置正确