package account

import "errors"

// 账户相关错误
var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInvalidAccountID     = errors.New("invalid account id")
	ErrEmptyEmail           = errors.New("email cannot be empty")
	ErrInvalidConfig        = errors.New("invalid configuration")
)