package email

import "errors"

// 领域层错误定义

var (
	// 邮件相关错误
	ErrMessageNotFound      = errors.New("message not found")
	ErrMessageAlreadyExists = errors.New("message already exists")
	ErrInvalidMessageID     = errors.New("invalid message id")
	ErrEmptySubject         = errors.New("subject cannot be empty")
	ErrEmptyBody            = errors.New("body cannot be empty")
	ErrEmptyRecipient       = errors.New("recipient cannot be empty")

	// 地址相关错误
	ErrInvalidEmailAddress  = errors.New("invalid email address")
	ErrEmptyAddress         = errors.New("email address is empty")

	// 附件相关错误
	ErrAttachmentNotFound   = errors.New("attachment not found")
	ErrAttachmentTooLarge   = errors.New("attachment too large")
	ErrInvalidAttachment    = errors.New("invalid attachment")

	// 文件夹相关错误
	ErrFolderNotFound       = errors.New("folder not found")
	ErrFolderAlreadyExists  = errors.New("folder already exists")
	ErrInvalidFolderName    = errors.New("invalid folder name")
	ErrCannotDeleteSystemFolder = errors.New("cannot delete system folder")
)

// DomainError 领域错误结构
type DomainError struct {
	Code    string
	Message string
	Err     error
}

// Error 实现error接口
func (e *DomainError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap 支持错误解包
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError 创建领域错误
func NewDomainError(code string, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}