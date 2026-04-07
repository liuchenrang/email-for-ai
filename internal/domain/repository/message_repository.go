package repository

import (
	"context"

	"github.com/chenji/email/internal/domain/email"
)

// MessageRepository 邕件仓储接口
type MessageRepository interface {
	// Save 保存邮件
	Save(ctx context.Context, message *email.Message) error

	// SaveBatch 批量保存邮件
	SaveBatch(ctx context.Context, messages []*email.Message) error

	// FindByID 根据ID查找邮件
	FindByID(ctx context.Context, id email.MessageID) (*email.Message, error)

	// FindByMessageID 根据原始Message-ID查找邮件
	FindByMessageID(ctx context.Context, messageID string) (*email.Message, error)

	// FindByAccount 查找账户下的所有邮件
	FindByAccount(ctx context.Context, accountID email.AccountID, limit int, offset int) ([]*email.Message, error)

	// FindByFolder 查找文件夹下的邮件
	FindByFolder(ctx context.Context, folderID email.FolderID, limit int, offset int) ([]*email.Message, error)

	// FindUnread 查找未读邮件
	FindUnread(ctx context.Context, accountID email.AccountID, limit int) ([]*email.Message, error)

	// Search 全文搜索邮件
	Search(ctx context.Context, query string, accountID email.AccountID, limit int) ([]*email.Message, error)

	// SearchWithFilter 带过滤条件的搜索
	SearchWithFilter(ctx context.Context, query string, filter SearchFilter) ([]*email.Message, error)

	// Update 更新邮件
	Update(ctx context.Context, message *email.Message) error

	// Delete 删除邮件
	Delete(ctx context.Context, id email.MessageID) error

	// DeleteByFolder 删除文件夹下的所有邮件
	DeleteByFolder(ctx context.Context, folderID email.FolderID) error

	// Count 统计邮件数量
	Count(ctx context.Context, accountID email.AccountID) (int64, error)

	// CountByFolder 统计文件夹邮件数量
	CountByFolder(ctx context.Context, folderID email.FolderID) (int64, error)

	// CountUnread 统计未读邮件数量
	CountUnread(ctx context.Context, accountID email.AccountID) (int64, error)

	// Exists 检查邮件是否存在
	Exists(ctx context.Context, id email.MessageID) (bool, error)

	// ExistsByMessageID 检查原始Message-ID是否存在
	ExistsByMessageID(ctx context.Context, messageID string) (bool, error)
}

// SearchFilter 搜索过滤条件
type SearchFilter struct {
	AccountID  email.AccountID
	FolderID   email.FolderID
	From       string
	To         string
	Subject    string
	Since      string
	Until      string
	IsUnread   bool
	HasAttachment bool
	Limit      int
	Offset     int
}

// AttachmentRepository 附件仓储接口
type AttachmentRepository interface {
	// Save 保存附件
	Save(ctx context.Context, attachment *email.Attachment, messageID email.MessageID) error

	// FindByID 根据ID查找附件
	FindByID(ctx context.Context, id email.AttachmentID) (*email.Attachment, error)

	// FindByMessage 查找邮件的所有附件
	FindByMessage(ctx context.Context, messageID email.MessageID) ([]*email.Attachment, error)

	// Delete 删除附件
	Delete(ctx context.Context, id email.AttachmentID) error

	// DeleteByMessage 删除邮件的所有附件
	DeleteByMessage(ctx context.Context, messageID email.MessageID) error
}