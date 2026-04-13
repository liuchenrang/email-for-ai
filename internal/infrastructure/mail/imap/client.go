package imap

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/email"
	"github.com/chenji/email/internal/domain/folder"
	imaplib "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// IDCommand IMAP ID命令 (RFC 2971)
// 网易邮箱(163/126/188)要求发送此命令才能正常操作
type IDCommand struct {
	ID map[string]string
}

// Command 实现imap.Commander接口
func (cmd *IDCommand) Command() *imaplib.Command {
	var fields []interface{}
	for k, v := range cmd.ID {
		fields = append(fields, k, v)
	}
	// ID 命令参数需要作为单个列表传递: ID ("key" "value" ...)
	return &imaplib.Command{
		Name:      "ID",
		Arguments: []interface{}{fields},
	}
}

// IMAPClient IMAP邮件客户端
type IMAPClient struct {
	client *client.Client
	config account.IMAPConfig
}

// NewIMAPClient 创建IMAP客户端
func NewIMAPClient(config account.IMAPConfig) *IMAPClient {
	return &IMAPClient{
		config: config,
	}
}

// Connect 连接到IMAP服务器
func (c *IMAPClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.config.Host(), c.config.Port())

	var cli *client.Client
	var err error

	// 163邮箱等需要使用TLS
	if c.config.Port() == 993 {
		cli, err = client.DialTLS(addr, nil)
	} else {
		cli, err = client.Dial(addr)
	}

	if err != nil {
		return fmt.Errorf("连接IMAP服务器失败: %w", err)
	}

	c.client = cli

	// 发送ID命令（网易邮箱必需，RFC 2971）
	// 必须在登录前发送，否则后续SELECT等操作会被拒绝
	idCmd := &IDCommand{
		ID: map[string]string{
			"name":    "email-cli",
			"version": "1.0",
			"vendor":  "local",
		},
	}
	if _, err := c.client.Execute(idCmd, nil); err != nil {
		// ID命令失败不阻止登录，某些服务器可能不支持
		// 但网易邮箱必须成功，所以记录警告但继续
		fmt.Printf("警告: ID命令发送失败: %v\n", err)
	}

	// 登录 - 使用用户名和授权码
	username := c.config.Username()
	// 如果用户名为空，说明配置有问题，但继续尝试（可能会失败）
	if err := c.client.Login(username, c.config.Password()); err != nil {
		return fmt.Errorf("登录失败 (用户: %s): %w", username, err)
	}

	return nil
}

// Disconnect 断开连接
func (c *IMAPClient) Disconnect() error {
	if c.client != nil {
		return c.client.Logout()
	}
	return nil
}

// ListFolders 列出所有文件夹
func (c *IMAPClient) ListFolders() ([]*folder.Folder, error) {
	mailboxes := make(chan *imaplib.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.client.List("", "*", mailboxes)
	}()

	folders := make([]*folder.Folder, 0)

	for mbox := range mailboxes {
		folderType := determineFolderType(mbox.Name)
		folders = append(folders, folder.NewFolder("", mbox.Name, folderType))
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取文件夹列表失败: %w", err)
	}

	return folders, nil
}

// FetchMessages 获取邮件
func (c *IMAPClient) FetchMessages(folderName string, limit int) ([]*email.Message, error) {
	// 选择文件夹
	mbox, err := c.client.Select(folderName, false)
	if err != nil {
		return nil, fmt.Errorf("选择文件夹失败 [%s]: %w", folderName, err)
	}

	if mbox.Messages == 0 {
		return []*email.Message{}, nil
	}

	// 计算获取范围
	from := uint32(1)
	to := mbox.Messages
	if limit > 0 && to > uint32(limit) {
		from = to - uint32(limit) + 1
	}

	// 创建获取范围
	seqset := new(imaplib.SeqSet)
	seqset.AddRange(from, to)

	// 获取邮件信封和正文
	items := []imaplib.FetchItem{
		imaplib.FetchEnvelope,
		imaplib.FetchFlags,
		imaplib.FetchInternalDate,
		imaplib.FetchRFC822, // 获取完整邮件内容
	}

	messages := make(chan *imaplib.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.client.Fetch(seqset, items, messages)
	}()

	result := make([]*email.Message, 0)

	for msg := range messages {
		if msg == nil || msg.Envelope == nil {
			continue
		}

		// 解析邮件（包括正文）
		m := parseIMAPMessage(msg, folderName)
		result = append(result, m)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件失败: %w", err)
	}

	return result, nil
}

// parseIMAPMessage 解析IMAP邮件为领域对象
func parseIMAPMessage(msg *imaplib.Message, folderName string) *email.Message {
	env := msg.Envelope

	// 解析发件人
	from := email.NewAddress("", "")
	if len(env.From) > 0 {
		from = email.NewAddress(env.From[0].PersonalName, env.From[0].Address())
	}

	// 解析收件人
	to := make([]email.Address, 0, len(env.To))
	for _, addr := range env.To {
		to = append(to, email.NewAddress(addr.PersonalName, addr.Address()))
	}

	// 解析抄送
	cc := make([]email.Address, 0, len(env.Cc))
	for _, addr := range env.Cc {
		cc = append(cc, email.NewAddress(addr.PersonalName, addr.Address()))
	}

	// 解析标记
	flags := email.NewFlags()
	for _, f := range msg.Flags {
		switch f {
		case imaplib.SeenFlag:
			flags = flags.Add(email.FlagSeen)
		case imaplib.FlaggedFlag:
			flags = flags.Add(email.FlagFlagged)
		case imaplib.DeletedFlag:
			flags = flags.Add(email.FlagDeleted)
		case imaplib.DraftFlag:
			flags = flags.Add(email.FlagDraft)
		}
	}

	// 提取邮件正文
	var rawBody string
	for _, literal := range msg.Body {
		if literal != nil {
			// 读取完整邮件内容
			bodyBytes, err := io.ReadAll(literal)
			if err == nil && len(bodyBytes) > 0 {
				rawBody = string(bodyBytes)
				break // 只需要第一个有效的内容
			}
		}
	}

	// 创建邮件对象
	body := email.NewBody(rawBody, "", "utf-8")
	date := msg.InternalDate
	if date.IsZero() {
		date = env.Date
	}

	return email.NewMessage(
		email.AccountID(""),     // 需要在上层设置
		email.FolderID(""),      // 需要在上层设置
		env.MessageId,
		env.Subject,
		from,
		to,
		date,
		body,
	)
}

// determineFolderType 根据文件夹名称确定类型
func determineFolderType(name string) folder.FolderType {
	nameLower := strings.ToLower(name)

	switch {
	case strings.Contains(nameLower, "inbox"):
		return folder.FolderTypeInbox
	case strings.Contains(nameLower, "sent"):
		return folder.FolderTypeSent
	case strings.Contains(nameLower, "draft") || strings.Contains(nameLower, "drafts"):
		return folder.FolderTypeDrafts
	case strings.Contains(nameLower, "trash") || strings.Contains(nameLower, "deleted"):
		return folder.FolderTypeTrash
	case strings.Contains(nameLower, "spam") || strings.Contains(nameLower, "junk"):
		return folder.FolderTypeSpam
	case strings.Contains(nameLower, "archive"):
		return folder.FolderTypeArchive
	default:
		return folder.FolderTypeCustom
	}
}

// 确保time包被使用
var _ = time.Second