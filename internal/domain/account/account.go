package account

import (
	"time"

	"github.com/google/uuid"
)

// AccountID 账户ID值对象
type AccountID string

// NewAccountID 创建新的账户ID
func NewAccountID() AccountID {
	return AccountID("acc_" + uuid.New().String()[:8])
}

// String 返回字符串表示
func (id AccountID) String() string {
	return string(id)
}

// IsEmpty 检查是否为空
func (id AccountID) IsEmpty() bool {
	return id == ""
}

// Account 账户聚合根
type Account struct {
	id          AccountID
	name        string
	email       string
	imapConfig  IMAPConfig
	smtpConfig  SMTPConfig
	pop3Config  *POP3Config
	isActive    bool
	isDefault   bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewAccount 创建新账户
func NewAccount(
	name string,
	email string,
	imapConfig IMAPConfig,
	smtpConfig SMTPConfig,
) *Account {
	now := time.Now()
	return &Account{
		id:         NewAccountID(),
		name:       name,
		email:      email,
		imapConfig: imapConfig,
		smtpConfig: smtpConfig,
		pop3Config: nil,
		isActive:   true,
		isDefault:  false,
		createdAt:  now,
		updatedAt:  now,
	}
}

// Reconstruction 从持久化重建账户
func Reconstruction(
	id AccountID,
	name string,
	email string,
	imapConfig IMAPConfig,
	smtpConfig SMTPConfig,
	pop3Config *POP3Config,
	isActive bool,
	isDefault bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Account {
	return &Account{
		id:         id,
		name:       name,
		email:      email,
		imapConfig: imapConfig,
		smtpConfig: smtpConfig,
		pop3Config: pop3Config,
		isActive:   isActive,
		isDefault:  isDefault,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Getters
func (a *Account) ID() AccountID         { return a.id }
func (a *Account) Name() string          { return a.name }
func (a *Account) Email() string         { return a.email }
func (a *Account) IMAPConfig() IMAPConfig { return a.imapConfig }
func (a *Account) SMTPConfig() SMTPConfig { return a.smtpConfig }
func (a *Account) POP3Config() *POP3Config { return a.pop3Config }
func (a *Account) IsActive() bool        { return a.isActive }
func (a *Account) IsDefault() bool       { return a.isDefault }
func (a *Account) CreatedAt() time.Time  { return a.createdAt }
func (a *Account) UpdatedAt() time.Time  { return a.updatedAt }

// SetPOP3Config 设置POP3配置
func (a *Account) SetPOP3Config(config POP3Config) {
	a.pop3Config = &config
	a.updatedAt = time.Now()
}

// SetActive 设置活跃状态
func (a *Account) SetActive(active bool) {
	a.isActive = active
	a.updatedAt = time.Now()
}

// SetDefault 设置为默认账户
func (a *Account) SetDefault(isDefault bool) {
	a.isDefault = isDefault
	a.updatedAt = time.Now()
}

// HasPOP3 是否配置了POP3
func (a *Account) HasPOP3() bool {
	return a.pop3Config != nil
}