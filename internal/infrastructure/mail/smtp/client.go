package smtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/email"
)

// SMTPClient SMTP邮件发送客户端
type SMTPClient struct {
	config account.SMTPConfig
}

// NewSMTPClient 创建SMTP客户端
func NewSMTPClient(config account.SMTPConfig) *SMTPClient {
	return &SMTPClient{
		config: config,
	}
}

// Send 发送邮件
func (c *SMTPClient) Send(msg *email.Message) error {
	// 构建邮件内容
	from := msg.From().Email()
	to := msg.To()
	toEmails := make([]string, len(to))
	for i, addr := range to {
		toEmails[i] = addr.Email()
	}

	// 添加抄送
	if len(msg.CC()) > 0 {
		for _, addr := range msg.CC() {
			toEmails = append(toEmails, addr.Email())
		}
	}

	// 构建邮件正文
	emailContent := c.buildEmailContent(msg)

	// 根据端口选择连接方式：
	// - 465 端口：直接 TLS（SMTPS）
	// - 587/25 端口：STARTTLS
	addr := fmt.Sprintf("%s:%d", c.config.Host(), c.config.Port())
	if c.config.Port() == 465 {
		return c.sendWithTLS(addr, from, toEmails, emailContent)
	}
	return c.sendWithSTARTTLS(addr, from, toEmails, emailContent)
}

// sendWithSTARTTLS 使用STARTTLS发送（适用于587/25端口）
func (c *SMTPClient) sendWithSTARTTLS(addr string, from string, to []string, content string) error {
	// 创建SMTP客户端（明文连接）
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}
	defer client.Close()

	// 发送EHLO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO失败: %w", err)
	}

	// 检查服务器是否支持STARTTLS
	ok, _ := client.Extension("STARTTLS")
	if ok {
		// 配置TLS
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // 允许自签名证书
			ServerName:         c.config.Host(),
		}
		// 升级到TLS连接
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS升级失败: %w", err)
		}
	}

	// 认证
	if c.config.HasAuth() {
		auth := smtp.PlainAuth("", c.config.Username(), c.config.Password(), c.config.Host())
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("认证失败: %w", err)
		}
	}

	// 设置发件人
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人失败 %s: %w", recipient, err)
		}
	}

	// 发送正文
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据写入器失败: %w", err)
	}

	_, err = w.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("写入内容失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭写入器失败: %w", err)
	}

	return client.Quit()
}

// sendWithTLS 使用TLS发送（适用于465端口）
func (c *SMTPClient) sendWithTLS(addr string, from string, to []string, content string) error {
	// 创建TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 允许自签名证书
		ServerName:         c.config.Host(),
	}

	// 使用TLS连接服务器
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, c.config.Host())
	if err != nil {
		conn.Close()
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	// 发送EHLO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO失败: %w", err)
	}

	// 认证
	if c.config.HasAuth() {
		auth := smtp.PlainAuth("", c.config.Username(), c.config.Password(), c.config.Host())
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("认证失败: %w", err)
		}
	}

	// 设置发件人
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人失败 %s: %w", recipient, err)
		}
	}

	// 发送正文
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据写入器失败: %w", err)
	}

	_, err = w.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("写入内容失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭写入器失败: %w", err)
	}

	return client.Quit()
}

// buildEmailContent 构建邮件内容
func (c *SMTPClient) buildEmailContent(msg *email.Message) string {
	var content bytes.Buffer

	// From
	fromDisplay := msg.From().String()
	content.WriteString(fmt.Sprintf("From: %s\r\n", fromDisplay))

	// To
	toStrs := make([]string, len(msg.To()))
	for i, addr := range msg.To() {
		toStrs[i] = addr.String()
	}
	content.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(toStrs, ", ")))

	// CC
	if len(msg.CC()) > 0 {
		ccStrs := make([]string, len(msg.CC()))
		for i, addr := range msg.CC() {
			ccStrs[i] = addr.String()
		}
		content.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(ccStrs, ", ")))
	}

	// Subject - Base64编码支持中文
	subject := msg.Subject()
	encodedSubject := base64.StdEncoding.EncodeToString([]byte(subject))
	content.WriteString(fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", encodedSubject))

	// Date
	content.WriteString(fmt.Sprintf("Date: %s\r\n", msg.Date().Format("Mon, 02 Jan 2006 15:04:05 -0700")))

	// MIME Headers
	content.WriteString("MIME-Version: 1.0\r\n")
	content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	content.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	content.WriteString("\r\n")

	// Body
	content.WriteString(msg.Body().Text())

	return content.String()
}