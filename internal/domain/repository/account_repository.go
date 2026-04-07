package repository

import (
	"context"

	"github.com/chenji/email/internal/domain/account"
)

// AccountRepository 账户仓储接口
type AccountRepository interface {
	// Save 保存账户
	Save(ctx context.Context, acc *account.Account) error

	// FindByID 根据ID查找账户
	FindByID(ctx context.Context, id account.AccountID) (*account.Account, error)

	// FindByEmail 根据邮箱地址查找账户
	FindByEmail(ctx context.Context, email string) (*account.Account, error)

	// FindAll 查找所有账户
	FindAll(ctx context.Context) ([]*account.Account, error)

	// FindActive 查找活跃账户
	FindActive(ctx context.Context) ([]*account.Account, error)

	// FindDefault 查找默认账户
	FindDefault(ctx context.Context) (*account.Account, error)

	// Update 更新账户
	Update(ctx context.Context, acc *account.Account) error

	// Delete 删除账户
	Delete(ctx context.Context, id account.AccountID) error

	// Exists 检查账户是否存在
	Exists(ctx context.Context, id account.AccountID) (bool, error)

	// ExistsByEmail 检查邮箱是否已存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// SetDefault 设置默认账户
	SetDefault(ctx context.Context, id account.AccountID) error

	// Count 统计账户数量
	Count(ctx context.Context) (int64, error)
}