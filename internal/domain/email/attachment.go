package email

import (
	"github.com/google/uuid"
)

// AttachmentID 附件ID值对象
type AttachmentID string

// NewAttachmentID 创建新的附件ID
func NewAttachmentID() AttachmentID {
	return AttachmentID("att_" + uuid.New().String()[:8])
}

// Attachment 附件实体
type Attachment struct {
	id          AttachmentID
	filename    string
	mimeType    string
	size        int64
	contentID   string
	storedPath  string
}

// NewAttachment 创建新的附件
func NewAttachment(filename string, mimeType string, size int64, contentID string) *Attachment {
	return &Attachment{
		id:         NewAttachmentID(),
		filename:   filename,
		mimeType:   mimeType,
		size:       size,
		contentID:  contentID,
	}
}

// Reconstruction 从持久化重建附件
func ReconstructAttachment(
	id AttachmentID,
	filename string,
	mimeType string,
	size int64,
	contentID string,
	storedPath string,
) *Attachment {
	return &Attachment{
		id:         id,
		filename:   filename,
		mimeType:   mimeType,
		size:       size,
		contentID:  contentID,
		storedPath: storedPath,
	}
}

// Getters
func (a *Attachment) ID() AttachmentID   { return a.id }
func (a *Attachment) Filename() string   { return a.filename }
func (a *Attachment) MimeType() string   { return a.mimeType }
func (a *Attachment) Size() int64        { return a.size }
func (a *Attachment) ContentID() string  { return a.contentID }
func (a *Attachment) StoredPath() string { return a.storedPath }

// SetStoredPath 设置存储路径
func (a *Attachment) SetStoredPath(path string) {
	a.storedPath = path
}

// IsInline 是否为内嵌附件
func (a *Attachment) IsInline() bool {
	return a.contentID != ""
}