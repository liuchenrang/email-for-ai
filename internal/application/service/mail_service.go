package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/chenji/email/internal/domain/email"
	"github.com/chenji/email/internal/domain/repository"
)

// MailService 邮件应用服务
type MailService struct {
	messageRepo repository.MessageRepository
}

// NewMailService 创建邮件服务
func NewMailService(messageRepo repository.MessageRepository) *MailService {
	return &MailService{
		messageRepo: messageRepo,
	}
}

// ListMessages 列出邮件
func (s *MailService) ListMessages(ctx context.Context, accountID string, folder string, unreadOnly bool, limit int) ([]*email.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	if unreadOnly {
		return s.messageRepo.FindUnread(ctx, email.AccountID(accountID), limit)
	}

	if folder != "" && folder != "INBOX" {
		// 根据文件夹名称查找
		// TODO: 实现文件夹名到ID的映射
	}

	return s.messageRepo.FindByAccount(ctx, email.AccountID(accountID), limit, 0)
}

// GetMessage 获取邮件详情
func (s *MailService) GetMessage(ctx context.Context, messageID string) (*email.Message, error) {
	return s.messageRepo.FindByID(ctx, email.MessageID(messageID))
}

// MarkAsRead 标记为已读
func (s *MailService) MarkAsRead(ctx context.Context, messageID string) error {
	msg, err := s.messageRepo.FindByID(ctx, email.MessageID(messageID))
	if err != nil {
		return err
	}
	msg.SetRead(true)
	return s.messageRepo.Update(ctx, msg)
}

// SearchMessages 搜索邮件
func (s *MailService) SearchMessages(ctx context.Context, accountID string, query string, limit int) ([]*email.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.messageRepo.Search(ctx, query, email.AccountID(accountID), limit)
}

// HookConfig 钩子配置
type HookConfig struct {
	Enabled  bool   `json:"enabled"`
	Command  string `json:"command"`  // 要执行的命令
	Timeout  int    `json:"timeout"`  // 超时时间（秒）
	OnNew    bool   `json:"on_new"`   // 新邮件触发
	OnRead   bool   `json:"on_read"`  // 阅读邮件触发
	OnSend   bool   `json:"on_send"`  // 发送邮件触发
}

// HookPayload 钩子负载数据
type HookPayload struct {
	Event     string      `json:"event"`      // 事件类型: new_mail, read_mail, send_mail
	Timestamp string      `json:"timestamp"`  // 时间戳
	AccountID string      `json:"account_id"` // 账户ID
	Message   MessageInfo `json:"message"`    // 邮件信息
}

// MessageInfo 邮件信息（用于钩子）
type MessageInfo struct {
	ID          string `json:"id"`
	MessageID   string `json:"message_id"`
	Subject     string `json:"subject"`
	FromName    string `json:"from_name"`
	FromEmail   string `json:"from_email"`
	To          string `json:"to"`
	Date        string `json:"date"`
	BodyPreview string `json:"body_preview"` // 正文预览（前500字符）
	IsRead      bool   `json:"is_read"`
}

// ExecuteHook 执行钩子
func (s *MailService) ExecuteHook(hook HookConfig, event string, msg *email.Message) error {
	if !hook.Enabled {
		return nil
	}

	// 检查事件类型
	switch event {
	case "new_mail":
		if !hook.OnNew {
			return nil
		}
	case "read_mail":
		if !hook.OnRead {
			return nil
		}
	case "send_mail":
		if !hook.OnSend {
			return nil
		}
	}

	// 构建负载数据
	toStrs := make([]string, len(msg.To()))
	for i, addr := range msg.To() {
		toStrs[i] = addr.Email()
	}

	bodyPreview := msg.Body().Text()
	if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500]
	}

	payload := HookPayload{
		Event:     event,
		Timestamp: time.Now().Format(time.RFC3339),
		AccountID: msg.AccountID().String(),
		Message: MessageInfo{
			ID:          msg.ID().String(),
			MessageID:   msg.MessageID(),
			Subject:     msg.Subject(),
			FromName:    msg.From().Name(),
			FromEmail:   msg.From().Email(),
			To:          strings.Join(toStrs, ", "),
			Date:        msg.Date().Format(time.RFC3339),
			BodyPreview: bodyPreview,
			IsRead:      msg.IsRead(),
		},
	}

	// 序列化为JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal hook payload: %w", err)
	}

	// 执行钩子命令
	cmd := exec.Command("sh", "-c", hook.Command)
	cmd.Stdin = strings.NewReader(string(payloadJSON))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook command failed: %w, output: %s", err, string(output))
	}

	return nil
}

// ExecuteHookRaw 直接传递JSON执行钩子
func (s *MailService) ExecuteHookRaw(command string, payloadJSON []byte) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = strings.NewReader(string(payloadJSON))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook command failed: %w, output: %s", err, string(output))
	}

	return nil
}