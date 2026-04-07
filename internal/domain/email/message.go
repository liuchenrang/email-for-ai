package email

import (
	"time"

	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/folder"
	"github.com/google/uuid"
)

// MessageID 邕件唯一标识值对象
type MessageID string

// NewMessageID 创建新的邮件ID
func NewMessageID() MessageID {
	return MessageID("msg_" + uuid.New().String()[:8])
}

// String 返回字符串表示
func (id MessageID) String() string {
	return string(id)
}

// IsEmpty 检查是否为空
func (id MessageID) IsEmpty() bool {
	return id == ""
}

// 使用其他模块的ID类型
type AccountID = account.AccountID
type FolderID = folder.FolderID

// Message 邕件聚合根
type Message struct {
	id          MessageID
	messageID   string       // 原始邮件 Message-ID header
	accountID   AccountID
	folderID    FolderID
	subject     string
	from        Address
	to          []Address
	cc          []Address
	date        time.Time
	body        Body
	attachments []Attachment
	flags       Flags
	size        int64
	isRead      bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewMessage 创建新的邮件实体
func NewMessage(
	accountID AccountID,
	folderID FolderID,
	messageID string,
	subject string,
	from Address,
	to []Address,
	date time.Time,
	body Body,
) *Message {
	now := time.Now()
	return &Message{
		id:          NewMessageID(),
		messageID:   messageID,
		accountID:   accountID,
		folderID:    folderID,
		subject:     subject,
		from:        from,
		to:          to,
		date:        date,
		body:        body,
		attachments: []Attachment{},
		flags:       Flags{},
		isRead:      false,
		createdAt:   now,
		updatedAt:   now,
	}
}

// Reconstruction 从持久化重建邮件实体
func Reconstruction(
	id MessageID,
	accountID AccountID,
	folderID FolderID,
	messageID string,
	subject string,
	from Address,
	to []Address,
	cc []Address,
	date time.Time,
	body Body,
	attachments []Attachment,
	flags Flags,
	size int64,
	isRead bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Message {
	return &Message{
		id:          id,
		messageID:   messageID,
		accountID:   accountID,
		folderID:    folderID,
		subject:     subject,
		from:        from,
		to:          to,
		cc:          cc,
		date:        date,
		body:        body,
		attachments: attachments,
		flags:       flags,
		size:        size,
		isRead:      isRead,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Getters - 获取属性值

func (m *Message) ID() MessageID           { return m.id }
func (m *Message) MessageID() string       { return m.messageID }
func (m *Message) AccountID() AccountID    { return m.accountID }
func (m *Message) FolderID() FolderID      { return m.folderID }
func (m *Message) Subject() string         { return m.subject }
func (m *Message) From() Address           { return m.from }
func (m *Message) To() []Address           { return m.to }
func (m *Message) CC() []Address           { return m.cc }
func (m *Message) Date() time.Time         { return m.date }
func (m *Message) Body() Body              { return m.body }
func (m *Message) Attachments() []Attachment { return m.attachments }
func (m *Message) Flags() Flags            { return m.flags }
func (m *Message) Size() int64             { return m.size }
func (m *Message) IsRead() bool            { return m.isRead }
func (m *Message) CreatedAt() time.Time    { return m.createdAt }
func (m *Message) UpdatedAt() time.Time    { return m.updatedAt }

// HasAttachments 是否有附件
func (m *Message) HasAttachments() bool {
	return len(m.attachments) > 0
}

// HasCC 是否有抄送
func (m *Message) HasCC() bool {
	return len(m.cc) > 0
}

// SetRead 设置已读状态
func (m *Message) SetRead(read bool) {
	m.isRead = read
	m.updatedAt = time.Now()
}

// AddAttachment 添加附件
func (m *Message) AddAttachment(att Attachment) {
	m.attachments = append(m.attachments, att)
	m.updatedAt = time.Now()
}

// AddFlag 添加标记
func (m *Message) AddFlag(flag Flag) {
	m.flags = m.flags.Add(flag)
	m.updatedAt = time.Now()
}

// RemoveFlag 移除标记
func (m *Message) RemoveFlag(flag Flag) {
	m.flags = m.flags.Remove(flag)
	m.updatedAt = time.Now()
}

// SetFolder 设置文件夹
func (m *Message) SetFolder(folderID FolderID) {
	m.folderID = folderID
	m.updatedAt = time.Now()
}

// SetCC 设置抄送
func (m *Message) SetCC(cc []Address) {
	m.cc = cc
	m.updatedAt = time.Now()
}

// SetSize 设置邮件大小
func (m *Message) SetSize(size int64) {
	m.size = size
	m.updatedAt = time.Now()
}