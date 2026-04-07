package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/chenji/email/internal/domain/folder"
)

// FolderRepository 文件夹仓储SQLite实现
type FolderRepository struct {
	db *DB
}

// NewFolderRepository 创建文件夹仓储
func NewFolderRepository(db *DB) *FolderRepository {
	return &FolderRepository{db: db}
}

// Save 保存文件夹
func (r *FolderRepository) Save(ctx context.Context, f *folder.Folder) error {
	query := `
		INSERT INTO folders (id, account_id, name, type, message_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(ctx, query,
		f.ID().String(),
		f.AccountID(),
		f.Name(),
		f.Type(),
		f.MessageCount(),
		f.CreatedAt(),
	)

	return err
}

// FindByID 根据ID查找文件夹
func (r *FolderRepository) FindByID(ctx context.Context, id folder.FolderID) (*folder.Folder, error) {
	query := `
		SELECT id, account_id, name, type, message_count, created_at
		FROM folders WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, id.String())
	return r.scanFolder(row)
}

// FindByName 根据名称查找文件夹
func (r *FolderRepository) FindByName(ctx context.Context, accountID string, name string) (*folder.Folder, error) {
	query := `
		SELECT id, account_id, name, type, message_count, created_at
		FROM folders WHERE account_id = ? AND name = ?
	`

	row := r.db.QueryRow(ctx, query, accountID, name)
	return r.scanFolder(row)
}

// FindByAccount 查找账户下的所有文件夹
func (r *FolderRepository) FindByAccount(ctx context.Context, accountID string) ([]*folder.Folder, error) {
	query := `
		SELECT id, account_id, name, type, message_count, created_at
		FROM folders WHERE account_id = ? ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFolders(rows)
}

// FindByType 根据类型查找文件夹
func (r *FolderRepository) FindByType(ctx context.Context, accountID string, folderType folder.FolderType) (*folder.Folder, error) {
	query := `
		SELECT id, account_id, name, type, message_count, created_at
		FROM folders WHERE account_id = ? AND type = ?
	`

	row := r.db.QueryRow(ctx, query, accountID, folderType)
	return r.scanFolder(row)
}

// Update 更新文件夹
func (r *FolderRepository) Update(ctx context.Context, f *folder.Folder) error {
	query := `
		UPDATE folders SET name = ?, type = ?, message_count = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(ctx, query,
		f.Name(),
		f.Type(),
		f.MessageCount(),
		f.ID().String(),
	)

	return err
}

// Delete 删除文件夹
func (r *FolderRepository) Delete(ctx context.Context, id folder.FolderID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM folders WHERE id = ?", id.String())
	return err
}

// Exists 检查文件夹是否存在
func (r *FolderRepository) Exists(ctx context.Context, id folder.FolderID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM folders WHERE id = ?", id.String()).Scan(&count)
	return count > 0, err
}

// ExistsByName 检查文件夹名称是否已存在
func (r *FolderRepository) ExistsByName(ctx context.Context, accountID string, name string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM folders WHERE account_id = ? AND name = ?", accountID, name).Scan(&count)
	return count > 0, err
}

// UpdateMessageCount 更新邮件计数
func (r *FolderRepository) UpdateMessageCount(ctx context.Context, id folder.FolderID, count int) error {
	_, err := r.db.Exec(ctx, "UPDATE folders SET message_count = ? WHERE id = ?", count, id.String())
	return err
}

// Count 统计文件夹数量
func (r *FolderRepository) Count(ctx context.Context, accountID string) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM folders WHERE account_id = ?", accountID).Scan(&count)
	return count, err
}

// scanFolder 扫描单个文件夹
func (r *FolderRepository) scanFolder(row *sql.Row) (*folder.Folder, error) {
	var (
		id, accountID, name, folderType string
		messageCount int
		createdAt time.Time
	)

	err := row.Scan(&id, &accountID, &name, &folderType, &messageCount, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, folder.ErrFolderNotFound
		}
		return nil, err
	}

	return folder.Reconstruction(
		folder.FolderID(id),
		accountID,
		name,
		folder.FolderType(folderType),
		messageCount,
		createdAt,
	), nil
}

// scanFolders 扫描多个文件夹
func (r *FolderRepository) scanFolders(rows *sql.Rows) ([]*folder.Folder, error) {
	var folders []*folder.Folder

	for rows.Next() {
		var (
			id, accountID, name, folderType string
			messageCount int
			createdAt time.Time
		)

		err := rows.Scan(&id, &accountID, &name, &folderType, &messageCount, &createdAt)
		if err != nil {
			return nil, err
		}

		folders = append(folders, folder.Reconstruction(
			folder.FolderID(id),
			accountID,
			name,
			folder.FolderType(folderType),
			messageCount,
			createdAt,
		))
	}

	return folders, nil
}