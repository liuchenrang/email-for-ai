package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/email"
	"github.com/chenji/email/internal/domain/folder"
	"github.com/chenji/email/internal/domain/repository"
	imapClient "github.com/chenji/email/internal/infrastructure/mail/imap"
	pop3Client "github.com/chenji/email/internal/infrastructure/mail/pop3"
)

// SyncService 邮件同步服务
type SyncService struct {
	accountRepo repository.AccountRepository
	messageRepo repository.MessageRepository
	folderRepo  repository.FolderRepository
}

// NewSyncService 创建同步服务
func NewSyncService(
	accountRepo repository.AccountRepository,
	messageRepo repository.MessageRepository,
	folderRepo repository.FolderRepository,
) *SyncService {
	return &SyncService{
		accountRepo: accountRepo,
		messageRepo: messageRepo,
		folderRepo:  folderRepo,
	}
}

// SyncResult 同步结果
type SyncResult struct {
	AccountID string
	Protocol  string
	Synced    int
	New       int
	Skipped   int
	Errors    []string
}

// Sync 同步指定账户的邮件
func (s *SyncService) Sync(ctx context.Context, accountID string, protocol string, limit int) (*SyncResult, error) {
	// 获取账户
	acc, err := s.accountRepo.FindByID(ctx, account.AccountID(accountID))
	if err != nil {
		return nil, fmt.Errorf("账户不存在: %w", err)
	}

	result := &SyncResult{
		AccountID: accountID,
		Protocol:  protocol,
		Errors:    []string{},
	}

	switch protocol {
	case "imap":
		return s.syncIMAP(ctx, acc, limit, result)
	case "pop3":
		return s.syncPOP3(ctx, acc, limit, result)
	default:
		return nil, fmt.Errorf("不支持的协议: %s", protocol)
	}
}

// syncIMAP 使用IMAP同步
func (s *SyncService) syncIMAP(ctx context.Context, acc *account.Account, limit int, result *SyncResult) (*SyncResult, error) {
	// 创建IMAP客户端
	client := imapClient.NewIMAPClient(acc.IMAPConfig())

	// 连接
	if err := client.Connect(); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}
	defer client.Disconnect()

	// 获取文件夹列表
	folders, err := client.ListFolders()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("获取文件夹列表失败: %v", err))
		return result, err
	}

	// 保存文件夹
	for _, f := range folders {
		// 设置账户ID
		newFolder := folder.NewFolder(acc.ID().String(), f.Name(), f.Type())
		if err := s.folderRepo.Save(ctx, newFolder); err != nil {
			// 文件夹可能已存在，忽略错误
		}
	}

	// 同步INBOX邮件
	inboxMessages, err := client.FetchMessages("INBOX", limit)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("获取邮件失败: %v", err))
		return result, err
	}

	// 保存邮件
	for _, msg := range inboxMessages {
		// 设置账户ID
		msg = setEmailAccountID(msg, acc.ID())

		// 检查是否已存在
		if msg.MessageID() != "" {
			exists, _ := s.messageRepo.ExistsByMessageID(ctx, msg.MessageID())
			if exists {
				result.Skipped++
				continue
			}
		}

		if err := s.messageRepo.Save(ctx, msg); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("保存邮件失败: %v", err))
			continue
		}

		result.Synced++
		result.New++
	}

	return result, nil
}

// syncPOP3 使用POP3同步
func (s *SyncService) syncPOP3(ctx context.Context, acc *account.Account, limit int, result *SyncResult) (*SyncResult, error) {
	if !acc.HasPOP3() {
		return nil, fmt.Errorf("账户未配置POP3")
	}

	// 创建POP3客户端
	client := pop3Client.NewPOP3Client(*acc.POP3Config())

	// 连接
	if err := client.Connect(); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}
	defer client.Disconnect()

	// 统计邮件数量
	count, _, err := client.CountMessages()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("统计邮件数量失败: %v", err))
		return result, err
	}

	// 计算获取范围
	start := 1
	if limit > 0 && count > limit {
		start = count - limit + 1
	}

	// 获取邮件
	for i := start; i <= count; i++ {
		msg, err := client.FetchMessage(i)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("获取邮件 %d 失败: %v", i, err))
			continue
		}

		// 设置账户ID
		msg = setEmailAccountID(msg, acc.ID())

		// 检查是否已存在（通过Message-ID判断）
		if msg.MessageID() != "" {
			exists, _ := s.messageRepo.ExistsByMessageID(ctx, msg.MessageID())
			if exists {
				result.Skipped++
				continue
			}
		}

		// 保存邮件
		if err := s.messageRepo.Save(ctx, msg); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("保存邮件 %d 失败: %v", i, err))
			continue
		}

		result.Synced++
		result.New++
	}

	return result, nil
}

// setEmailAccountID 设置邮件的账户ID（创建新对象）
func setEmailAccountID(msg *email.Message, accountID account.AccountID) *email.Message {
	return email.Reconstruction(
		msg.ID(),
		accountID,
		msg.FolderID(),
		msg.MessageID(),
		msg.Subject(),
		msg.From(),
		msg.To(),
		msg.CC(),
		msg.Date(),
		msg.Body(),
		msg.Attachments(),
		msg.Flags(),
		msg.Size(),
		msg.IsRead(),
		msg.CreatedAt(),
		msg.UpdatedAt(),
	)
}

// SyncAll 同步所有活跃账户
func (s *SyncService) SyncAll(ctx context.Context, protocol string, limit int) ([]*SyncResult, error) {
	accounts, err := s.accountRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取活跃账户失败: %w", err)
	}

	results := make([]*SyncResult, 0, len(accounts))
	for _, acc := range accounts {
		result, err := s.Sync(ctx, acc.ID().String(), protocol, limit)
		if err != nil {
			results = append(results, &SyncResult{
				AccountID: acc.ID().String(),
				Protocol:  protocol,
				Errors:    []string{err.Error()},
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// 确保time包被使用
var _ = time.Second