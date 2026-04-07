package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/chenji/email/internal/domain/email"
	"github.com/chenji/email/internal/domain/repository"
)

// MessageRepository 邕件仓储SQLite实现
type MessageRepository struct {
	db *DB
}

// NewMessageRepository 创建邮件仓储
func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Save 保存邮件
func (r *MessageRepository) Save(ctx context.Context, message *email.Message) error {
	query := `
		INSERT INTO messages (
			id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	toJSON, _ := json.Marshal(message.To())
	ccJSON, _ := json.Marshal(message.CC())
	flagsJSON, _ := json.Marshal(message.Flags().Strings())

	_, err := r.db.Exec(ctx, query,
		message.ID().String(),
		message.AccountID().String(),
		message.FolderID().String(),
		message.MessageID(),
		message.Subject(),
		message.From().Name(),
		message.From().Email(),
		string(toJSON),
		string(ccJSON),
		message.Date(),
		message.Body().Text(),
		message.Body().HTML(),
		string(flagsJSON),
		message.Size(),
		message.IsRead(),
		message.CreatedAt(),
		message.UpdatedAt(),
	)

	return err
}

// SaveBatch 批量保存邮件
func (r *MessageRepository) SaveBatch(ctx context.Context, messages []*email.Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, m := range messages {
		if err := r.saveInTx(tx, m); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// saveInTx 在事务中保存邮件
func (r *MessageRepository) saveInTx(tx *sql.Tx, message *email.Message) error {
	query := `
		INSERT INTO messages (
			id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	toJSON, _ := json.Marshal(message.To())
	ccJSON, _ := json.Marshal(message.CC())
	flagsJSON, _ := json.Marshal(message.Flags().Strings())

	_, err := tx.Exec(query,
		message.ID().String(),
		message.AccountID().String(),
		message.FolderID().String(),
		message.MessageID(),
		message.Subject(),
		message.From().Name(),
		message.From().Email(),
		string(toJSON),
		string(ccJSON),
		message.Date(),
		message.Body().Text(),
		message.Body().HTML(),
		string(flagsJSON),
		message.Size(),
		message.IsRead(),
		message.CreatedAt(),
		message.UpdatedAt(),
	)

	return err
}

// FindByID 根据ID查找邮件
func (r *MessageRepository) FindByID(ctx context.Context, id email.MessageID) (*email.Message, error) {
	query := `
		SELECT id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		FROM messages WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, id.String())
	return r.scanMessage(row)
}

// FindByMessageID 根据原始Message-ID查找邮件
func (r *MessageRepository) FindByMessageID(ctx context.Context, messageID string) (*email.Message, error) {
	query := `
		SELECT id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		FROM messages WHERE message_id = ?
	`

	row := r.db.QueryRow(ctx, query, messageID)
	return r.scanMessage(row)
}

// FindByAccount 查找账户下的所有邮件
func (r *MessageRepository) FindByAccount(ctx context.Context, accountID email.AccountID, limit int, offset int) ([]*email.Message, error) {
	query := `
		SELECT id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		FROM messages WHERE account_id = ?
		ORDER BY date DESC LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, query, accountID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// FindByFolder 查找文件夹下的邮件
func (r *MessageRepository) FindByFolder(ctx context.Context, folderID email.FolderID, limit int, offset int) ([]*email.Message, error) {
	query := `
		SELECT id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		FROM messages WHERE folder_id = ?
		ORDER BY date DESC LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, query, folderID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// FindUnread 查找未读邮件
func (r *MessageRepository) FindUnread(ctx context.Context, accountID email.AccountID, limit int) ([]*email.Message, error) {
	query := `
		SELECT id, account_id, folder_id, message_id, subject,
			from_name, from_email, to_addrs, cc_addrs, date,
			text_body, html_body, flags, size, is_read,
			created_at, updated_at
		FROM messages WHERE account_id = ? AND is_read = 0
		ORDER BY date DESC LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, accountID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Search 全文搜索邮件
func (r *MessageRepository) Search(ctx context.Context, query string, accountID email.AccountID, limit int) ([]*email.Message, error) {
	sqlQuery := `
		SELECT m.id, m.account_id, m.folder_id, m.message_id, m.subject,
			m.from_name, m.from_email, m.to_addrs, m.cc_addrs, m.date,
			m.text_body, m.html_body, m.flags, m.size, m.is_read,
			m.created_at, m.updated_at
		FROM messages m
		JOIN messages_fts fts ON m.rowid = fts.rowid
		WHERE messages_fts MATCH ? AND m.account_id = ?
		ORDER BY date DESC LIMIT ?
	`

	rows, err := r.db.Query(ctx, sqlQuery, query, accountID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// SearchWithFilter 带过滤条件的搜索
func (r *MessageRepository) SearchWithFilter(ctx context.Context, query string, filter repository.SearchFilter) ([]*email.Message, error) {
	baseQuery := `
		SELECT m.id, m.account_id, m.folder_id, m.message_id, m.subject,
			m.from_name, m.from_email, m.to_addrs, m.cc_addrs, m.date,
			m.text_body, m.html_body, m.flags, m.size, m.is_read,
			m.created_at, m.updated_at
		FROM messages m
	`

	var conditions []string
	var args []interface{}

	if query != "" {
		baseQuery += " JOIN messages_fts fts ON m.rowid = fts.rowid "
		conditions = append(conditions, "messages_fts MATCH ?")
		args = append(args, query)
	}

	if !filter.AccountID.IsEmpty() {
		conditions = append(conditions, "m.account_id = ?")
		args = append(args, filter.AccountID.String())
	}

	if !filter.FolderID.IsEmpty() {
		conditions = append(conditions, "m.folder_id = ?")
		args = append(args, filter.FolderID.String())
	}

	if filter.From != "" {
		conditions = append(conditions, "m.from_email LIKE ?")
		args = append(args, "%"+filter.From+"%")
	}

	if filter.To != "" {
		conditions = append(conditions, "m.to_addrs LIKE ?")
		args = append(args, "%"+filter.To+"%")
	}

	if filter.Subject != "" {
		conditions = append(conditions, "m.subject LIKE ?")
		args = append(args, "%"+filter.Subject+"%")
	}

	if filter.Since != "" {
		conditions = append(conditions, "m.date >= ?")
		args = append(args, filter.Since)
	}

	if filter.Until != "" {
		conditions = append(conditions, "m.date <= ?")
		args = append(args, filter.Until)
	}

	if filter.IsUnread {
		conditions = append(conditions, "m.is_read = 0")
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE "
		for i, cond := range conditions {
			if i > 0 {
				baseQuery += " AND "
			}
			baseQuery += cond
		}
	}

	baseQuery += " ORDER BY m.date DESC "

	if filter.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Update 更新邮件
func (r *MessageRepository) Update(ctx context.Context, message *email.Message) error {
	query := `
		UPDATE messages SET
			folder_id = ?, subject = ?, from_name = ?, from_email = ?,
			to_addrs = ?, cc_addrs = ?, date = ?, text_body = ?, html_body = ?,
			flags = ?, size = ?, is_read = ?, updated_at = ?
		WHERE id = ?
	`

	toJSON, _ := json.Marshal(message.To())
	ccJSON, _ := json.Marshal(message.CC())
	flagsJSON, _ := json.Marshal(message.Flags().Strings())

	_, err := r.db.Exec(ctx, query,
		message.FolderID().String(),
		message.Subject(),
		message.From().Name(),
		message.From().Email(),
		string(toJSON),
		string(ccJSON),
		message.Date(),
		message.Body().Text(),
		message.Body().HTML(),
		string(flagsJSON),
		message.Size(),
		message.IsRead(),
		time.Now(),
		message.ID().String(),
	)

	return err
}

// Delete 删除邮件
func (r *MessageRepository) Delete(ctx context.Context, id email.MessageID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM messages WHERE id = ?", id.String())
	return err
}

// DeleteByFolder 删除文件夹下的所有邮件
func (r *MessageRepository) DeleteByFolder(ctx context.Context, folderID email.FolderID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM messages WHERE folder_id = ?", folderID.String())
	return err
}

// Count 统计邮件数量
func (r *MessageRepository) Count(ctx context.Context, accountID email.AccountID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE account_id = ?", accountID.String()).Scan(&count)
	return count, err
}

// CountByFolder 统计文件夹邮件数量
func (r *MessageRepository) CountByFolder(ctx context.Context, folderID email.FolderID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE folder_id = ?", folderID.String()).Scan(&count)
	return count, err
}

// CountUnread 统计未读邮件数量
func (r *MessageRepository) CountUnread(ctx context.Context, accountID email.AccountID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE account_id = ? AND is_read = 0", accountID.String()).Scan(&count)
	return count, err
}

// Exists 检查邮件是否存在
func (r *MessageRepository) Exists(ctx context.Context, id email.MessageID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE id = ?", id.String()).Scan(&count)
	return count > 0, err
}

// ExistsByMessageID 检查原始Message-ID是否存在
func (r *MessageRepository) ExistsByMessageID(ctx context.Context, messageID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE message_id = ?", messageID).Scan(&count)
	return count > 0, err
}

// scanMessage 扫描单个邮件
func (r *MessageRepository) scanMessage(row *sql.Row) (*email.Message, error) {
	var (
		id, accountID, folderID, messageID, subject string
		fromName, fromEmail, toJSON, ccJSON         string
		textBody, htmlBody, flagsJSON               string
		size                                        int64
		isRead                                      int
		date, createdAt, updatedAt                  time.Time
	)

	err := row.Scan(
		&id, &accountID, &folderID, &messageID, &subject,
		&fromName, &fromEmail, &toJSON, &ccJSON, &date,
		&textBody, &htmlBody, &flagsJSON, &size, &isRead,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, email.ErrMessageNotFound
		}
		return nil, err
	}

	// 解析JSON字段
	var toAddrs, ccAddrs []email.Address
	json.Unmarshal([]byte(toJSON), &toAddrs)
	json.Unmarshal([]byte(ccJSON), &ccAddrs)

	var flagStrs []string
	json.Unmarshal([]byte(flagsJSON), &flagStrs)
	flags := email.NewFlags()
	for _, s := range flagStrs {
		flags = flags.Add(email.Flag(s))
	}

	from := email.NewAddress(fromName, fromEmail)
	body := email.NewBody(textBody, htmlBody, "utf-8")

	return email.Reconstruction(
		email.MessageID(id),
		email.AccountID(accountID),
		email.FolderID(folderID),
		messageID,
		subject,
		from,
		toAddrs,
		ccAddrs,
		date,
		body,
		[]email.Attachment{},
		flags,
		size,
		isRead == 1,
		createdAt,
		updatedAt,
	), nil
}

// scanMessages 扫描多个邮件
func (r *MessageRepository) scanMessages(rows *sql.Rows) ([]*email.Message, error) {
	var messages []*email.Message

	for rows.Next() {
		var (
			id, accountID, folderID, messageID, subject string
			fromName, fromEmail, toJSON, ccJSON         string
			textBody, htmlBody, flagsJSON               string
			size                                        int64
			isRead                                      int
			date, createdAt, updatedAt                  time.Time
		)

		err := rows.Scan(
			&id, &accountID, &folderID, &messageID, &subject,
			&fromName, &fromEmail, &toJSON, &ccJSON, &date,
			&textBody, &htmlBody, &flagsJSON, &size, &isRead,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		var toAddrs, ccAddrs []email.Address
		json.Unmarshal([]byte(toJSON), &toAddrs)
		json.Unmarshal([]byte(ccJSON), &ccAddrs)

		var flagStrs []string
		json.Unmarshal([]byte(flagsJSON), &flagStrs)
		flags := email.NewFlags()
		for _, s := range flagStrs {
			flags = flags.Add(email.Flag(s))
		}

		from := email.NewAddress(fromName, fromEmail)
		body := email.NewBody(textBody, htmlBody, "utf-8")

		messages = append(messages, email.Reconstruction(
			email.MessageID(id),
			email.AccountID(accountID),
			email.FolderID(folderID),
			messageID,
			subject,
			from,
			toAddrs,
			ccAddrs,
			date,
			body,
			[]email.Attachment{},
			flags,
			size,
			isRead == 1,
			createdAt,
			updatedAt,
		))
	}

	return messages, nil
}