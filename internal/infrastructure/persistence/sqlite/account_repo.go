package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/chenji/email/internal/domain/account"
)

// AccountRepository 账户仓储SQLite实现
type AccountRepository struct {
	db *DB
}

// NewAccountRepository 创建账户仓储
func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Save 保存账户
func (r *AccountRepository) Save(ctx context.Context, acc *account.Account) error {
	query := `
		INSERT INTO accounts (
			id, name, email, imap_host, imap_port, imap_user, imap_pass,
			smtp_host, smtp_port, smtp_user, smtp_pass,
			pop3_host, pop3_port, pop3_user, pop3_pass,
			is_active, is_default, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	pop3Host, pop3Port, pop3User, pop3Pass := "", 0, "", ""
	if acc.HasPOP3() {
		pop3Host = acc.POP3Config().Host()
		pop3Port = acc.POP3Config().Port()
		pop3User = acc.POP3Config().Username()
		pop3Pass = acc.POP3Config().Password()
	}

	_, err := r.db.Exec(ctx, query,
		acc.ID().String(),
		acc.Name(),
		acc.Email(),
		acc.IMAPConfig().Host(),
		acc.IMAPConfig().Port(),
		acc.IMAPConfig().Username(),
		acc.IMAPConfig().Password(),
		acc.SMTPConfig().Host(),
		acc.SMTPConfig().Port(),
		acc.SMTPConfig().Username(),
		acc.SMTPConfig().Password(),
		pop3Host, pop3Port, pop3User, pop3Pass,
		acc.IsActive(),
		acc.IsDefault(),
		acc.CreatedAt(),
		acc.UpdatedAt(),
	)

	return err
}

// FindByID 根据ID查找账户
func (r *AccountRepository) FindByID(ctx context.Context, id account.AccountID) (*account.Account, error) {
	query := `
		SELECT id, name, email, imap_host, imap_port, imap_user, imap_pass,
			smtp_host, smtp_port, smtp_user, smtp_pass,
			pop3_host, pop3_port, pop3_user, pop3_pass,
			is_active, is_default, created_at, updated_at
		FROM accounts WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, id.String())
	return r.scanAccount(row)
}

// FindByEmail 根据邮箱地址查找账户
func (r *AccountRepository) FindByEmail(ctx context.Context, email string) (*account.Account, error) {
	query := `
		SELECT id, name, email, imap_host, imap_port, imap_user, imap_pass,
			smtp_host, smtp_port, smtp_user, smtp_pass,
			pop3_host, pop3_port, pop3_user, pop3_pass,
			is_active, is_default, created_at, updated_at
		FROM accounts WHERE email = ?
	`

	row := r.db.QueryRow(ctx, query, email)
	return r.scanAccount(row)
}

// FindAll 查找所有账户
func (r *AccountRepository) FindAll(ctx context.Context) ([]*account.Account, error) {
	query := `
		SELECT id, name, email, imap_host, imap_port, imap_user, imap_pass,
			smtp_host, smtp_port, smtp_user, smtp_pass,
			pop3_host, pop3_port, pop3_user, pop3_pass,
			is_active, is_default, created_at, updated_at
		FROM accounts ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAccounts(rows)
}

// FindActive 查找活跃账户
func (r *AccountRepository) FindActive(ctx context.Context) ([]*account.Account, error) {
	query := `
		SELECT id, name, email, imap_host, imap_port, imap_user, imap_pass,
			smtp_host, smtp_port, smtp_user, smtp_pass,
			pop3_host, pop3_port, pop3_user, pop3_pass,
			is_active, is_default, created_at, updated_at
		FROM accounts WHERE is_active = 1 ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAccounts(rows)
}

// FindDefault 查找默认账户
func (r *AccountRepository) FindDefault(ctx context.Context) (*account.Account, error) {
	query := `
		SELECT id, name, email, imap_host, imap_port, imap_user, imap_pass,
			smtp_host, smtp_port, smtp_user, smtp_pass,
			pop3_host, pop3_port, pop3_user, pop3_pass,
			is_active, is_default, created_at, updated_at
		FROM accounts WHERE is_default = 1
	`

	row := r.db.QueryRow(ctx, query)
	return r.scanAccount(row)
}

// Update 更新账户
func (r *AccountRepository) Update(ctx context.Context, acc *account.Account) error {
	query := `
		UPDATE accounts SET
			name = ?, email = ?, imap_host = ?, imap_port = ?, imap_user = ?, imap_pass = ?,
			smtp_host = ?, smtp_port = ?, smtp_user = ?, smtp_pass = ?,
			pop3_host = ?, pop3_port = ?, pop3_user = ?, pop3_pass = ?,
			is_active = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`

	pop3Host, pop3Port, pop3User, pop3Pass := "", 0, "", ""
	if acc.HasPOP3() {
		pop3Host = acc.POP3Config().Host()
		pop3Port = acc.POP3Config().Port()
		pop3User = acc.POP3Config().Username()
		pop3Pass = acc.POP3Config().Password()
	}

	_, err := r.db.Exec(ctx, query,
		acc.Name(),
		acc.Email(),
		acc.IMAPConfig().Host(),
		acc.IMAPConfig().Port(),
		acc.IMAPConfig().Username(),
		acc.IMAPConfig().Password(),
		acc.SMTPConfig().Host(),
		acc.SMTPConfig().Port(),
		acc.SMTPConfig().Username(),
		acc.SMTPConfig().Password(),
		pop3Host, pop3Port, pop3User, pop3Pass,
		acc.IsActive(),
		acc.IsDefault(),
		time.Now(),
		acc.ID().String(),
	)

	return err
}

// Delete 删除账户
func (r *AccountRepository) Delete(ctx context.Context, id account.AccountID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM accounts WHERE id = ?", id.String())
	return err
}

// Exists 检查账户是否存在
func (r *AccountRepository) Exists(ctx context.Context, id account.AccountID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM accounts WHERE id = ?", id.String()).Scan(&count)
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否已存在
func (r *AccountRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM accounts WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

// SetDefault 设置默认账户
func (r *AccountRepository) SetDefault(ctx context.Context, id account.AccountID) error {
	// 先清除所有默认账户
	_, err := r.db.Exec(ctx, "UPDATE accounts SET is_default = 0")
	if err != nil {
		return err
	}

	// 设置指定账户为默认
	_, err = r.db.Exec(ctx, "UPDATE accounts SET is_default = 1 WHERE id = ?", id.String())
	return err
}

// Count 统计账户数量
func (r *AccountRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM accounts").Scan(&count)
	return count, err
}

// scanAccount 扫描单个账户
func (r *AccountRepository) scanAccount(row *sql.Row) (*account.Account, error) {
	var (
		id, name, email string
		imapHost, imapUser, imapPass string
		smtpHost, smtpUser, smtpPass string
		pop3Host, pop3User, pop3Pass sql.NullString
		imapPort, smtpPort int
		pop3Port sql.NullInt32
		isActive, isDefault int
		createdAt, updatedAt time.Time
	)

	err := row.Scan(
		&id, &name, &email, &imapHost, &imapPort, &imapUser, &imapPass,
		&smtpHost, &smtpPort, &smtpUser, &smtpPass,
		&pop3Host, &pop3Port, &pop3User, &pop3Pass,
		&isActive, &isDefault, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}

	imapConfig := account.NewIMAPConfig(imapHost, imapPort, imapUser, imapPass, imapPort == 993)
	smtpConfig := account.NewSMTPConfig(smtpHost, smtpPort, smtpUser, smtpPass, smtpPort == 465)

	var pop3Config *account.POP3Config
	if pop3Host.Valid && pop3Host.String != "" {
		pop3Config = &account.POP3Config{}
		*pop3Config = account.NewPOP3Config(
			pop3Host.String,
			int(pop3Port.Int32),
			pop3User.String,
			pop3Pass.String,
			int(pop3Port.Int32) == 995,
		)
	}

	return account.Reconstruction(
		account.AccountID(id),
		name,
		email,
		imapConfig,
		smtpConfig,
		pop3Config,
		isActive == 1,
		isDefault == 1,
		createdAt,
		updatedAt,
	), nil
}

// scanAccounts 扫描多个账户
func (r *AccountRepository) scanAccounts(rows *sql.Rows) ([]*account.Account, error) {
	var accounts []*account.Account

	for rows.Next() {
		var (
			id, name, email string
			imapHost, imapUser, imapPass string
			smtpHost, smtpUser, smtpPass string
			pop3Host, pop3User, pop3Pass sql.NullString
			imapPort, smtpPort int
			pop3Port sql.NullInt32
			isActive, isDefault int
			createdAt, updatedAt time.Time
		)

		err := rows.Scan(
			&id, &name, &email, &imapHost, &imapPort, &imapUser, &imapPass,
			&smtpHost, &smtpPort, &smtpUser, &smtpPass,
			&pop3Host, &pop3Port, &pop3User, &pop3Pass,
			&isActive, &isDefault, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		imapConfig := account.NewIMAPConfig(imapHost, imapPort, imapUser, imapPass, imapPort == 993)
		smtpConfig := account.NewSMTPConfig(smtpHost, smtpPort, smtpUser, smtpPass, smtpPort == 465)

		var pop3Config *account.POP3Config
		if pop3Host.Valid && pop3Host.String != "" {
			pop3Config = &account.POP3Config{}
			*pop3Config = account.NewPOP3Config(
				pop3Host.String,
				int(pop3Port.Int32),
				pop3User.String,
				pop3Pass.String,
				int(pop3Port.Int32) == 995,
			)
		}

		accounts = append(accounts, account.Reconstruction(
			account.AccountID(id),
			name,
			email,
			imapConfig,
			smtpConfig,
			pop3Config,
			isActive == 1,
			isDefault == 1,
			createdAt,
			updatedAt,
		))
	}

	return accounts, nil
}