package container

import (
	"sync"

	"github.com/chenji/email/internal/application/service"
	"github.com/chenji/email/internal/infrastructure/persistence/sqlite"
	"github.com/spf13/viper"
)

var (
	instance *Container
	once     sync.Once
)

// Container 依赖注入容器
type Container struct {
	db *sqlite.DB

	accountService *service.AccountService
	mailService    *service.MailService
	syncService    *service.SyncService
}

// GetContainer 获取容器单例
func GetContainer() *Container {
	once.Do(func() {
		instance = &Container{}
	})
	return instance
}

// InitDB 初始化数据库
func (c *Container) InitDB() error {
	dbPath := viper.GetString("database.path")
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return err
	}
	c.db = db
	return nil
}

// GetDB 获取数据库连接
func (c *Container) GetDB() *sqlite.DB {
	return c.db
}

// GetAccountService 获取账户服务
func (c *Container) GetAccountService() *service.AccountService {
	if c.accountService == nil {
		c.accountService = service.NewAccountService(
			sqlite.NewAccountRepository(c.db),
		)
	}
	return c.accountService
}

// GetMailService 获取邮件服务
func (c *Container) GetMailService() *service.MailService {
	if c.mailService == nil {
		c.mailService = service.NewMailService(
			sqlite.NewMessageRepository(c.db),
		)
	}
	return c.mailService
}

// GetSyncService 获取同步服务
func (c *Container) GetSyncService() *service.SyncService {
	if c.syncService == nil {
		c.syncService = service.NewSyncService(
			sqlite.NewAccountRepository(c.db),
			sqlite.NewMessageRepository(c.db),
			sqlite.NewFolderRepository(c.db),
		)
	}
	return c.syncService
}

// Close 关闭资源
func (c *Container) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}