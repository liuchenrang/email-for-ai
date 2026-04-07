package folder

import (
	"time"

	"github.com/google/uuid"
)

// FolderID 文件夹ID值对象
type FolderID string

// NewFolderID 创建新的文件夹ID
func NewFolderID() FolderID {
	return FolderID("fld_" + uuid.New().String()[:8])
}

// String 返回字符串表示
func (id FolderID) String() string {
	return string(id)
}

// IsEmpty 检查是否为空
func (id FolderID) IsEmpty() bool {
	return id == ""
}

// FolderType 文件夹类型枚举
type FolderType string

const (
	FolderTypeInbox    FolderType = "inbox"
	FolderTypeSent     FolderType = "sent"
	FolderTypeDrafts   FolderType = "drafts"
	FolderTypeTrash    FolderType = "trash"
	FolderTypeSpam     FolderType = "spam"
	FolderTypeArchive  FolderType = "archive"
	FolderTypeCustom   FolderType = "custom"
)

// IsSystemFolder 是否为系统文件夹
func (t FolderType) IsSystemFolder() bool {
	return t != FolderTypeCustom
}

// Folder 文件夹实体
type Folder struct {
	id           FolderID
	accountID    string
	name         string
	folderType   FolderType
	messageCount int
	createdAt    time.Time
}

// NewFolder 创建新文件夹
func NewFolder(accountID string, name string, folderType FolderType) *Folder {
	return &Folder{
		id:           NewFolderID(),
		accountID:    accountID,
		name:         name,
		folderType:   folderType,
		messageCount: 0,
		createdAt:    time.Now(),
	}
}

// Reconstruction 从持久化重建文件夹
func Reconstruction(
	id FolderID,
	accountID string,
	name string,
	folderType FolderType,
	messageCount int,
	createdAt time.Time,
) *Folder {
	return &Folder{
		id:           id,
		accountID:    accountID,
		name:         name,
		folderType:   folderType,
		messageCount: messageCount,
		createdAt:    createdAt,
	}
}

// Getters
func (f *Folder) ID() FolderID           { return f.id }
func (f *Folder) AccountID() string      { return f.accountID }
func (f *Folder) Name() string           { return f.name }
func (f *Folder) Type() FolderType       { return f.folderType }
func (f *Folder) MessageCount() int      { return f.messageCount }
func (f *Folder) CreatedAt() time.Time   { return f.createdAt }

// IsSystem 是否为系统文件夹
func (f *Folder) IsSystem() bool {
	return f.folderType.IsSystemFolder()
}

// UpdateMessageCount 更新邮件数量
func (f *Folder) UpdateMessageCount(count int) {
	f.messageCount = count
}

// IncrementMessageCount 增加邮件数量
func (f *Folder) IncrementMessageCount() {
	f.messageCount++
}

// DecrementMessageCount 减少邮件数量
func (f *Folder) DecrementMessageCount() {
	if f.messageCount > 0 {
		f.messageCount--
	}
}