package service

import (
	"context"
	"fmt"

	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/repository"
)

// AccountService 账户应用服务
type AccountService struct {
	accountRepo repository.AccountRepository
}

// NewAccountService 创建账户服务
func NewAccountService(accountRepo repository.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

// AddAccountInput 添加账户输入
type AddAccountInput struct {
	Name      string
	Email     string
	IMAPHost  string
	IMAPPort  int
	IMAPUser  string
	IMAPPass  string
	SMTPHost  string
	SMTPPort  int
	SMTPUser  string
	SMTPPass  string
	POP3Host  string
	POP3Port  int
	POP3User  string
	POP3Pass  string
}

// AddAccount 添加账户
func (s *AccountService) AddAccount(ctx context.Context, input AddAccountInput) (*account.Account, error) {
	// 检查邮箱是否已存在
	exists, err := s.accountRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("account with email %s already exists", input.Email)
	}

	// 创建IMAP配置
	imapConfig := account.NewIMAPConfig(
		input.IMAPHost,
		input.IMAPPort,
		input.IMAPUser,
		input.IMAPPass,
		input.IMAPPort == 993, // TLS for port 993
	)

	// 创建SMTP配置
	smtpConfig := account.NewSMTPConfig(
		input.SMTPHost,
		input.SMTPPort,
		input.SMTPUser,
		input.SMTPPass,
		input.SMTPPort == 465, // TLS for port 465
	)

	// 创建账户
	acc := account.NewAccount(input.Name, input.Email, imapConfig, smtpConfig)

	// 如果提供了POP3配置，设置它
	if input.POP3Host != "" {
		pop3Config := account.NewPOP3Config(
			input.POP3Host,
			input.POP3Port,
			input.POP3User,
			input.POP3Pass,
			input.POP3Port == 995, // TLS for port 995
		)
		acc.SetPOP3Config(pop3Config)
	}

	// 如果是第一个账户，设为默认
	count, err := s.accountRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count accounts: %w", err)
	}
	if count == 0 {
		acc.SetDefault(true)
	}

	// 保存账户
	if err := s.accountRepo.Save(ctx, acc); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return acc, nil
}

// ListAccounts 列出所有账户
func (s *AccountService) ListAccounts(ctx context.Context) ([]*account.Account, error) {
	return s.accountRepo.FindAll(ctx)
}

// GetAccount 获取账户
func (s *AccountService) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	return s.accountRepo.FindByID(ctx, account.AccountID(id))
}

// GetDefaultAccount 获取默认账户
func (s *AccountService) GetDefaultAccount(ctx context.Context) (*account.Account, error) {
	return s.accountRepo.FindDefault(ctx)
}

// RemoveAccount 删除账户
func (s *AccountService) RemoveAccount(ctx context.Context, id string) error {
	return s.accountRepo.Delete(ctx, account.AccountID(id))
}

// SetDefaultAccount 设置默认账户
func (s *AccountService) SetDefaultAccount(ctx context.Context, id string) error {
	return s.accountRepo.SetDefault(ctx, account.AccountID(id))
}