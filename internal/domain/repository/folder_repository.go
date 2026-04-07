package repository

import (
	"context"

	"github.com/chenji/email/internal/domain/folder"
)

// FolderRepository 文件夹仓储接口
type FolderRepository interface {
	// Save 保存文件夹
	Save(ctx context.Context, f *folder.Folder) error

	// FindByID 根据ID查找文件夹
	FindByID(ctx context.Context, id folder.FolderID) (*folder.Folder, error)

	// FindByName 根据名称查找文件夹
	FindByName(ctx context.Context, accountID string, name string) (*folder.Folder, error)

	// FindByAccount 查找账户下的所有文件夹
	FindByAccount(ctx context.Context, accountID string) ([]*folder.Folder, error)

	// FindByType 根据类型查找文件夹
	FindByType(ctx context.Context, accountID string, folderType folder.FolderType) (*folder.Folder, error)

	// Update 更新文件夹
	Update(ctx context.Context, f *folder.Folder) error

	// Delete 删除文件夹
	Delete(ctx context.Context, id folder.FolderID) error

	// Exists 检查文件夹是否存在
	Exists(ctx context.Context, id folder.FolderID) (bool, error)

	// ExistsByName 检查文件夹名称是否已存在
	ExistsByName(ctx context.Context, accountID string, name string) (bool, error)

	// UpdateMessageCount 更新邮件计数
	UpdateMessageCount(ctx context.Context, id folder.FolderID, count int) error

	// Count 统计文件夹数量
	Count(ctx context.Context, accountID string) (int64, error)
}