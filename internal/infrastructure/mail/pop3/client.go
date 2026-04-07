package pop3

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/email"
)

// POP3Client POP3邮件客户端
type POP3Client struct {
	config account.POP3Config
	conn   net.Conn
	reader *bufio.Reader
}

// NewPOP3Client 创建POP3客户端
func NewPOP3Client(config account.POP3Config) *POP3Client {
	return &POP3Client{
		config: config,
	}
}

// Connect 连接到POP3服务器
func (c *POP3Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.config.Host(), c.config.Port())

	var conn net.Conn
	var err error

	// 端口995使用TLS直接连接
	if c.config.Port() == 995 {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.config.Host(),
		}
		conn, err = tls.Dial("tcp", addr, tlsConfig)
	} else {
		conn, err = net.Dial("tcp", addr)
	}

	if err != nil {
		return fmt.Errorf("连接POP3服务器失败: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)

	// 读取欢迎消息
	_, err = c.readLine()
	if err != nil {
		return fmt.Errorf("读取欢迎消息失败: %w", err)
	}

	// 登录
	if c.config.HasAuth() {
		if err := c.login(); err != nil {
			return err
		}
	}

	return nil
}

// readLine 读取一行响应
func (c *POP3Client) readLine() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}

// sendCommand 发送命令并读取响应
func (c *POP3Client) sendCommand(cmd string) (string, error) {
	_, err := fmt.Fprintf(c.conn, "%s\r\n", cmd)
	if err != nil {
		return "", err
	}
	return c.readLine()
}

// login POP3登录
func (c *POP3Client) login() error {
	// USER命令
	resp, err := c.sendCommand(fmt.Sprintf("USER %s", c.config.Username()))
	if err != nil {
		return fmt.Errorf("USER命令失败: %w", err)
	}
	if strings.HasPrefix(resp, "-ERR") {
		return fmt.Errorf("USER命令被拒绝: %s", resp)
	}

	// PASS命令
	resp, err = c.sendCommand(fmt.Sprintf("PASS %s", c.config.Password()))
	if err != nil {
		return fmt.Errorf("PASS命令失败: %w", err)
	}
	if strings.HasPrefix(resp, "-ERR") {
		return fmt.Errorf("PASS命令被拒绝: %s", resp)
	}

	return nil
}

// Disconnect 断开连接
func (c *POP3Client) Disconnect() error {
	if c.conn != nil {
		c.sendCommand("QUIT")
		return c.conn.Close()
	}
	return nil
}

// CountMessages 统计邮件数量
func (c *POP3Client) CountMessages() (int, int, error) {
	resp, err := c.sendCommand("STAT")
	if err != nil {
		return 0, 0, fmt.Errorf("STAT命令失败: %w", err)
	}

	if !strings.HasPrefix(resp, "+OK") {
		return 0, 0, fmt.Errorf("STAT命令被拒绝: %s", resp)
	}

	// 解析响应 +OK count size
	parts := strings.Fields(resp)
	if len(parts) < 3 {
		return 0, 0, fmt.Errorf("无效的STAT响应: %s", resp)
	}

	count := 0
	size := 0
	fmt.Sscanf(parts[1], "%d", &count)
	fmt.Sscanf(parts[2], "%d", &size)

	return count, size, nil
}

// ListMessages 列出所有邮件
func (c *POP3Client) ListMessages() ([]int, []int, error) {
	resp, err := c.sendCommand("LIST")
	if err != nil {
		return nil, nil, fmt.Errorf("LIST命令失败: %w", err)
	}

	if !strings.HasPrefix(resp, "+OK") {
		return nil, nil, fmt.Errorf("LIST命令被拒绝: %s", resp)
	}

	// 读取多行响应
	msgNums := []int{}
	msgSizes := []int{}

	for {
		line, err := c.readLine()
		if err != nil {
			return nil, nil, err
		}

		if line == "." {
			break
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			num := 0
			size := 0
			fmt.Sscanf(parts[0], "%d", &num)
			fmt.Sscanf(parts[1], "%d", &size)
			msgNums = append(msgNums, num)
			msgSizes = append(msgSizes, size)
		}
	}

	return msgNums, msgSizes, nil
}

// FetchMessage 获取指定邮件
func (c *POP3Client) FetchMessage(msgNum int) (*email.Message, error) {
	resp, err := c.sendCommand(fmt.Sprintf("RETR %d", msgNum))
	if err != nil {
		return nil, fmt.Errorf("RETR命令失败: %w", err)
	}

	if !strings.HasPrefix(resp, "+OK") {
		return nil, fmt.Errorf("RETR命令被拒绝: %s", resp)
	}

	// 读取邮件内容
	var content strings.Builder
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}

		if line == "." {
			break
		}

		// POP3协议中.开头的行需要去掉一个点
		if strings.HasPrefix(line, "..") {
			line = line[1:]
		}

		content.WriteString(line)
		content.WriteString("\r\n")
	}

	// 解析邮件
	return parsePOP3Message(content.String())
}

// parsePOP3Message 解析POP3邮件
func parsePOP3Message(content string) (*email.Message, error) {
	// 邮件头和正文分离 - 使用CRLF或LF
	var headerEnd int
	if idx := strings.Index(content, "\r\n\r\n"); idx != -1 {
		headerEnd = idx
	} else if idx := strings.Index(content, "\n\n"); idx != -1 {
		headerEnd = idx
	}

	var header, body string
	if headerEnd > 0 {
		header = content[:headerEnd]
		// 跳过空行
		body = content[headerEnd:]
		body = strings.TrimPrefix(body, "\r\n\r\n")
		body = strings.TrimPrefix(body, "\n\n")
	} else {
		header = content
		body = ""
	}

	// 解析邮件头
	var subject, fromEmail, fromName, toEmail, dateStr, messageID string

	// 处理CRLF或LF
	header = strings.ReplaceAll(header, "\r\n", "\n")
	headerLines := strings.Split(header, "\n")

	for _, line := range headerLines {
		if line == "" {
			continue
		}

		// 处理折叠行（以空格或制表符开头的行是上一行的续行）
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			continue // 简化处理，跳过续行
		}

		colon := strings.Index(line, ":")
		if colon == -1 {
			continue
		}

		name := strings.TrimSpace(line[:colon])
		value := strings.TrimSpace(line[colon+1:])

		switch strings.ToLower(name) {
		case "subject":
			subject = value
		case "from":
			fromEmail = value
			// 尝试解析 "Name <email>" 格式
			if strings.Contains(value, "<") {
				idx := strings.Index(value, "<")
				fromName = strings.TrimSpace(value[:idx])
				fromEmail = strings.TrimSuffix(strings.TrimPrefix(value[idx:], "<"), ">")
			}
		case "to":
			toEmail = value
		case "date":
			dateStr = value
		case "message-id":
			// 移除尖括号
			messageID = strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">")
		}
	}

	// 解析日期
	var msgDate time.Time
	if dateStr != "" {
		// 清理日期字符串：移除括号内的时区名称如(CST)
		cleanDate := dateStr
		if idx := strings.Index(cleanDate, "("); idx > 0 {
			cleanDate = strings.TrimSpace(cleanDate[:idx])
		}

		// 尝试多种日期格式（Go的2可以匹配1-2位数字，02只匹配2位数字）
		formats := []string{
			"Mon, 2 Jan 2006 15:04:05 -0700",     // Fri, 3 Apr 2026 17:38:20 +0800
			"Mon, 02 Jan 2006 15:04:05 -0700",    // Fri, 03 Apr 2026 17:38:20 +0800
			"Mon, _2 Jan 2006 15:04:05 -0700",    // 支持空格填充的日期
			"2 Jan 2006 15:04:05 -0700",          // 无星期前缀
			"02 Jan 2006 15:04:05 -0700",
			time.RFC1123Z,
			time.RFC1123,
		}
		for _, format := range formats {
			var err error
			msgDate, err = time.Parse(format, cleanDate)
			if err == nil {
				break
			}
		}
	}
	if msgDate.IsZero() {
		msgDate = time.Now()
	}

	from := email.NewAddress(fromName, fromEmail)
	to := []email.Address{email.NewAddress("", toEmail)}
	emailBody := email.NewTextBody(strings.TrimSpace(body))

	return email.NewMessage(
		email.AccountID(""), // 需要在上层设置
		email.FolderID(""),  // POP3没有文件夹概念
		messageID,           // 从邮件头解析Message-ID
		subject,
		from,
		to,
		msgDate,
		emailBody,
	), nil
}