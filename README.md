# Email CLI Client

一个基于DDD架构的命令行邮件客户端，支持IMAP/POP3收件和SMTP发件。

## 功能特性

- 📥 **邮件收取**: 支持 IMAP 和 POP3 协议
- 📤 **邮件发送**: 支持 SMTP 协议（支持TLS/SSL）
- 🔍 **全文搜索**: 基于 SQLite FTS5 的本地搜索
- 💾 **本地存储**: 所有邮件数据存储在本地 SQLite
- 🤖 **AI友好**: 支持 JSON 输出格式，方便 AI 工具调用
- 🔔 **Hook支持**: 新邮件到达时可执行自定义脚本
- 📅 **正确日期**: 显示邮件实际发送时间，而非同步时间

## 安装

```bash
go build -o bin/email ./cmd/email
# 或
make build
```

## 快速开始

### 1. 添加账户

**163邮箱示例**（需要使用授权码，不是登录密码）：

```bash
./bin/email account add \
  --name "163邮箱" \
  --email "xxxxx@163.com" \
  --imap-host "imap.163.com" \
  --imap-port 993 \
  --imap-user "xxxxx@163.com" \
  --imap-pass "你的授权码" \
  --smtp-host "smtp.163.com" \
  --smtp-port 465 \
  --smtp-user "xxxxx@163.com" \
  --smtp-pass "你的授权码" \
  --pop3-host "pop.163.com" \
  --pop3-port 995 \
  --pop3-user "xxxxx@163.com" \
  --pop3-pass "你的授权码"
```

> **注意**: 163邮箱需要在网页版设置中开启"POP3/SMTP/IMAP"服务，获取授权码。

### 2. 同步邮件

```bash
# 使用IMAP同步（推荐，支持文件夹）
./bin/email sync --protocol imap --limit 50

# 使用POP3同步（163邮箱推荐，避免"Unsafe Login"问题）
./bin/email sync --protocol pop3 --limit 50

# 输出JSON格式
./bin/email sync --protocol pop3 --limit 50 -o json
```

**同步结果示例**:
```json
{
  "success": true,
  "data": {
    "account_id": "acc_02e9d71c",
    "new": 2,
    "skipped": 0,
    "synced": 2,
    "protocol": "pop3"
  }
}
```

### 3. 列出邮件

```bash
# 默认列出INBOX邮件
./bin/email list

# 仅显示未读邮件
./bin/email list --unread

# 限制显示数量
./bin/email list --limit 10

# JSON格式输出（方便AI解析）
./bin/email list -o json
```

**输出示例**:
```
      ID       |    SUBJECT     |      FROM       |      DATE      | STATUS | ATTACH
---------------+----------------+------------------+-----------------+--------+--------
  msg_44e963ad | xxxxx        | xxxxx         | 2026-04-03 17:52 | unread |
  msg_879718c8 | 新设备登录提醒 | 网易邮箱账号安全 | 2026-04-03 17:38 | unread |
```

### 4. 阅读邮件

```bash
# 阅读邮件内容
./bin/email read msg_44e963ad

# 显示原始邮件内容
./bin/email read msg_44e963ad --raw

# JSON格式输出
./bin/email read msg_44e963ad -o json
```

### 5. 发送邮件

```bash
# 直接发送
./bin/email compose \
  --to "recipient@example.com" \
  --subject "测试邮件" \
  --body "这是邮件正文内容" \
| ./bin/email send

# 或分开执行
./bin/email compose -t "xxx@xxxxx.com" -s "测试邮件" -b "正文" -o json
./bin/email send
```

### 6. 搜索邮件

```bash
# 关键词搜索
./bin/email search "报警"

# 按发件人搜索
./bin/email search --from "safe@service.netease.com"

# 按时间范围搜索
./bin/email search --since "2026-04-01"

# 组合搜索
./bin/email search "ecs" --from "Hangzhou" --since "2026-04-01" -o json
```

## 命令完整列表

| 命令 | 说明 | 示例 |
|------|------|------|
| `account add` | 添加邮件账户 | `email account add --name "163" ...` |
| `account list` | 列出所有账户 | `email account list` |
| `account remove` | 删除账户 | `email account remove <account-id>` |
| `account set-default` | 设置默认账户 | `email account set-default <account-id>` |
| `sync` | 同步邮件到本地 | `email sync --protocol pop3 --limit 50` |
| `list` | 列出本地邮件 | `email list --unread --limit 20` |
| `read` | 阅读邮件详情 | `email read <message-id>` |
| `search` | 全文搜索邮件 | `email search "关键词" -o json` |
| `compose` | 撰写邮件 | `email compose -t "xxx" -s "主题" -b "正文"` |
| `send` | 发送邮件 | `email send` |
| `folders list` | 列出文件夹 | `email folders list` |

## 全局选项

```
  -a, --account string   指定账户ID（默认使用第一个账户）
  -c, --config string    配置文件路径（默认 ~/.email/config.yaml）
  -o, --output string    输出格式: table|json（默认 table）
```

## 配置文件

默认路径: `~/.email/config.yaml`

```yaml
database:
  path: ~/.email/data.db

sync:
  protocol: pop3    # 默认同步协议
  limit: 100        # 默认同步数量

display:
  format: table     # 默认显示格式

hooks:
  on_new_email: "/path/to/script.sh"  # 新邮件到达时执行的脚本
```

## Hook 功能

当配置了 `hooks.on_new_email`，同步发现新邮件时会执行指定脚本，传入JSON数据：

```bash
# ~/.email/hooks/new_email.sh
#!/bin/bash
JSON_DATA=$(cat)
echo "收到新邮件: $(echo $JSON_DATA | jq -r '.subject')"
```

**传入的JSON格式**:
```json
{
  "account_id": "acc_02e9d71c",
  "message_id": "msg_44e963ad",
  "subject": "ecs报警",
  "from": {"name": "Hangzhou", "email": "Hangzhou@yehwang.com"},
  "date": "2026-04-03T17:52:15+08:00",
  "has_attachments": false
}
```

## 常见邮箱配置

### 163邮箱

```bash
--imap-host imap.163.com --imap-port 993
--smtp-host smtp.163.com --smtp-port 465
--pop3-host pop.163.com --pop3-port 995
```

> 使用POP3协议可避免IMAP的"Unsafe Login"错误

### QQ邮箱

```bash
--imap-host imap.qq.com --imap-port 993
--smtp-host smtp.qq.com --smtp-port 465
--pop3-host pop.qq.com --pop3-port 995
```

### Gmail

```bash
--imap-host imap.gmail.com --imap-port 993
--smtp-host smtp.gmail.com --smtp-port 587
```

## 数据存储

- 数据库: `~/.email/data.db` (SQLite)
- 支持FTS5全文搜索
- Message-ID去重，避免重复同步
- MIME头自动解码（Base64/Quoted-Printable）

## 架构设计

本项目采用领域驱动设计（DDD）架构：

```
internal/
├── domain/          # 领域层：聚合根(Message, Account)、值对象、领域服务
├── application/     # 应用层：应用服务(MailService, SyncService)
├── infrastructure/  # 基础设施层：仓储实现、IMAP/POP3/SMTP客户端
├── interface/       # 接口层：CLI命令
pkg/mime/            # MIME解码工具包
```

## 开发命令

```bash
# 安装依赖
go mod tidy

# 构建
go build -o bin/email ./cmd/email

# 清除本地数据
sqlite3 ~/.email/data.db "DELETE FROM messages;"
```