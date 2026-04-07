package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // 纯Go SQLite驱动
)

// DB SQLite数据库连接
type DB struct {
	conn *sql.DB
	path string
}

// NewDB 创建新的数据库连接
func NewDB(dbPath string) (*DB, error) {
	// 展开波浪号为用户目录
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[2:])
	}

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接参数
	conn.SetMaxOpenConns(1) // SQLite只支持单个连接
	conn.SetMaxIdleConns(1)

	// 验证连接
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// 执行迁移
	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Conn 获取数据库连接
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Path 获取数据库路径
func (db *DB) Path() string {
	return db.path
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// BeginTx 开始事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, opts)
}

// Exec 执行SQL语句
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}

// Query 执行查询
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

// QueryRow 执行单行查询
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRowContext(ctx, query, args...)
}

// Migrate 执行数据库迁移
func (db *DB) Migrate() error {
	migrations := getMigrations()

	for _, m := range migrations {
		if err := db.applyMigration(m); err != nil {
			return err
		}
	}

	return nil
}

// applyMigration 应用单个迁移
func (db *DB) applyMigration(m Migration) error {
	// 检查迁移是否已执行
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM migrations WHERE name = ?", m.Name,
	).Scan(&count)
	if err != nil {
		// migrations表不存在，创建它
		_, err = db.conn.Exec(`
			CREATE TABLE IF NOT EXISTS migrations (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create migrations table: %w", err)
		}
	}

	if count > 0 {
		return nil // 已执行
	}

	// 执行迁移
	_, err = db.conn.Exec(m.SQL)
	if err != nil {
		return fmt.Errorf("failed to apply migration %s: %w", m.Name, err)
	}

	// 记录迁移
	_, err = db.conn.Exec(
		"INSERT INTO migrations (name) VALUES (?)", m.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", m.Name, err)
	}

	return nil
}

// Migration 迁移定义
type Migration struct {
	Name string
	SQL  string
}

// getMigrations 获取所有迁移
func getMigrations() []Migration {
	return []Migration{
		{
			Name: "001_init",
			SQL: `
-- 账户表
CREATE TABLE IF NOT EXISTS accounts (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    email       TEXT NOT NULL UNIQUE,
    imap_host   TEXT NOT NULL,
    imap_port   INTEGER NOT NULL,
    imap_user   TEXT,
    imap_pass   TEXT,
    smtp_host   TEXT NOT NULL,
    smtp_port   INTEGER NOT NULL,
    smtp_user   TEXT,
    smtp_pass   TEXT,
    pop3_host   TEXT,
    pop3_port   INTEGER,
    pop3_user   TEXT,
    pop3_pass   TEXT,
    is_active   INTEGER DEFAULT 1,
    is_default  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 文件夹表
CREATE TABLE IF NOT EXISTS folders (
    id           TEXT PRIMARY KEY,
    account_id   TEXT NOT NULL,
    name         TEXT NOT NULL,
    type         TEXT DEFAULT 'custom',
    message_count INTEGER DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

-- 邕件表
CREATE TABLE IF NOT EXISTS messages (
    id           TEXT PRIMARY KEY,
    account_id   TEXT NOT NULL,
    folder_id    TEXT,
    message_id   TEXT,
    subject      TEXT,
    from_name    TEXT,
    from_email   TEXT NOT NULL,
    to_addrs     TEXT,
    cc_addrs     TEXT,
    date         DATETIME,
    text_body    TEXT,
    html_body    TEXT,
    flags        TEXT,
    size         INTEGER,
    is_read      INTEGER DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id),
    FOREIGN KEY (folder_id) REFERENCES folders(id)
);

-- FTS5全文搜索虚拟表
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    subject,
    from_name,
    from_email,
    to_addrs,
    text_body,
    content='messages',
    content_rowid='rowid'
);

-- 附件表
CREATE TABLE IF NOT EXISTS attachments (
    id           TEXT PRIMARY KEY,
    message_id   TEXT NOT NULL,
    filename     TEXT NOT NULL,
    mime_type    TEXT,
    size         INTEGER,
    content_id   TEXT,
    stored_path  TEXT,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_messages_account ON messages(account_id);
CREATE INDEX IF NOT EXISTS idx_messages_folder ON messages(folder_id);
CREATE INDEX IF NOT EXISTS idx_messages_date ON messages(date DESC);
CREATE INDEX IF NOT EXISTS idx_messages_read ON messages(is_read);
CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_folders_account ON folders(account_id);
			`,
		},
		{
			Name: "002_add_triggers",
			SQL: `
-- FTS5同步触发器：插入
CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, subject, from_name, from_email, to_addrs, text_body)
    VALUES (new.rowid, new.subject, new.from_name, new.from_email, new.to_addrs, new.text_body);
END;

-- FTS5同步触发器：删除
CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, subject, from_name, from_email, to_addrs, text_body)
    VALUES ('delete', old.rowid, old.subject, old.from_name, old.from_email, old.to_addrs, old.text_body);
END;

-- FTS5同步触发器：更新
CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, subject, from_name, from_email, to_addrs, text_body)
    VALUES ('delete', old.rowid, old.subject, old.from_name, old.from_email, old.to_addrs, old.text_body);
    INSERT INTO messages_fts(rowid, subject, from_name, from_email, to_addrs, text_body)
    VALUES (new.rowid, new.subject, new.from_name, new.from_email, new.to_addrs, new.text_body);
END;
			`,
		},
	}
}