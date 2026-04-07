package folder

import "errors"

// 文件夹相关错误
var (
	ErrFolderNotFound      = errors.New("folder not found")
	ErrFolderAlreadyExists = errors.New("folder already exists")
	ErrInvalidFolderID     = errors.New("invalid folder id")
	ErrInvalidFolderName   = errors.New("invalid folder name")
	ErrCannotDeleteSystem  = errors.New("cannot delete system folder")
)