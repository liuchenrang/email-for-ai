package email

import (
	"fmt"
	"strings"
)

// Address 邮件地址值对象
type Address struct {
	name  string
	email string
}

// NewAddress 创建新的邮件地址
func NewAddress(name string, email string) Address {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)
	return Address{
		name:  name,
		email: email,
	}
}

// ParseAddress 从字符串解析邮件地址
// 支持格式: "John Doe <john@example.com>" 或 "john@example.com"
func ParseAddress(addrStr string) Address {
	addrStr = strings.TrimSpace(addrStr)

	// 尝试解析 "Name <email>" 格式
	if strings.Contains(addrStr, "<") && strings.Contains(addrStr, ">") {
		start := strings.Index(addrStr, "<")
		end := strings.Index(addrStr, ">")
		if start < end {
			name := strings.TrimSpace(addrStr[:start])
			email := strings.TrimSpace(addrStr[start+1:end])
			return NewAddress(name, email)
		}
	}

	// 简单邮箱地址
	return NewAddress("", addrStr)
}

// Getters
func (a Address) Name() string  { return a.name }
func (a Address) Email() string { return a.email }

// String 返回完整地址字符串
func (a Address) String() string {
	if a.name == "" {
		return a.email
	}
	return fmt.Sprintf("%s <%s>", a.name, a.email)
}

// IsEmpty 检查是否为空
func (a Address) IsEmpty() bool {
	return a.email == ""
}

// Equals 检查两个地址是否相等
func (a Address) Equals(other Address) bool {
	return strings.EqualFold(a.email, other.email)
}

// Addresses 地址列表类型
type Addresses []Address

// String 返回地址列表的字符串表示
func (aa Addresses) String() string {
	strs := make([]string, len(aa))
	for i, a := range aa {
		strs[i] = a.String()
	}
	return strings.Join(strs, ", ")
}

// ParseAddresses 从字符串列表解析多个地址
func ParseAddresses(addrStrs []string) Addresses {
	result := make(Addresses, 0, len(addrStrs))
	for _, s := range addrStrs {
		if s != "" {
			result = append(result, ParseAddress(s))
		}
	}
	return result
}

// Emails 返回所有邮箱地址
func (aa Addresses) Emails() []string {
	emails := make([]string, len(aa))
	for i, a := range aa {
		emails[i] = a.email
	}
	return emails
}