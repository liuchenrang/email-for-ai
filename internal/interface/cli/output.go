package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// quiet 静默模式标志（从 root.go 设置）
var quiet bool

// SetQuietMode 设置静默模式
func SetQuietMode(q bool) {
	quiet = q
}

// isQuietMode 检查是否为静默模式
func isQuietMode() bool {
	return quiet
}

// Output 输出结果结构
type Output struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MetaInfo 元信息
type MetaInfo struct {
	Total   int `json:"total,omitempty"`
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

// NewSuccessOutput 创建成功输出
func NewSuccessOutput(data interface{}) *Output {
	return &Output{
		Success: true,
		Data:    data,
	}
}

// NewSuccessOutputWithMeta 创建带元信息的成功输出
func NewSuccessOutputWithMeta(data interface{}, total int, page int, perPage int) *Output {
	return &Output{
		Success: true,
		Data:    data,
		Meta: &MetaInfo{
			Total:   total,
			Page:    page,
			PerPage: perPage,
		},
	}
}

// NewErrorOutput 创建错误输出
func NewErrorOutput(code string, message string) *Output {
	return &Output{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// Print 打印输出，返回是否成功
func (o *Output) Print() bool {
	// 静默模式下成功不输出
	if isQuietMode() && o.Success {
		return true
	}

	if isJSONOutput() {
		o.PrintJSON()
	} else {
		o.PrintTable()
	}
	return o.Success
}

// PrintAndExit 打印输出并根据结果设置退出码
func (o *Output) PrintAndExit() {
	if !o.Print() {
		os.Exit(1)
	}
}

// PrintJSON 打印JSON格式
func (o *Output) PrintJSON() {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		fmt.Printf("{\"success\":false,\"error\":{\"code\":\"output_error\",\"message\":\"%s\"}}\n", err.Error())
		return
	}
	fmt.Println(string(data))
}

// PrintTable 打印表格格式
func (o *Output) PrintTable() {
	if !o.Success {
		color.Red("Error: %s (%s)", o.Error.Message, o.Error.Code)
		return
	}

	// 根据数据类型打印不同格式
	switch v := o.Data.(type) {
	case []MessageListItem:
		printMessageListTable(v)
	case *MessageDetail:
		printMessageDetailTable(v)
	case []AccountListItem:
		printAccountListTable(v)
	case []FolderListItem:
		printFolderListTable(v)
	case *SearchResult:
		printSearchResultTable(v)
	case string:
		fmt.Println(v)
	default:
		// 其他类型打印JSON
		o.PrintJSON()
	}
}

// MessageListItem 邕件列表项
type MessageListItem struct {
	ID        string `json:"id"`
	Subject   string `json:"subject"`
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
	Date      string `json:"date"`
	IsRead    bool   `json:"is_read"`
	HasAttach bool   `json:"has_attachments"`
}

// MessageDetail 邕件详情
type MessageDetail struct {
	ID          string           `json:"id"`
	MessageID   string           `json:"message_id,omitempty"`
	Subject     string           `json:"subject"`
	From        AddressInfo      `json:"from"`
	To          []AddressInfo    `json:"to"`
	CC          []AddressInfo    `json:"cc,omitempty"`
	Date        string           `json:"date"`
	Body        string           `json:"body"`
	Attachments []AttachmentInfo `json:"attachments,omitempty"`
	Flags       []string         `json:"flags"`
	IsRead      bool             `json:"is_read"`
}

// AddressInfo 地址信息
type AddressInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AttachmentInfo 附件信息
type AttachmentInfo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// AccountListItem 账户列表项
type AccountListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	IsDefault bool   `json:"is_default"`
}

// FolderListItem 文件夹列表项
type FolderListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	MessageCount int    `json:"message_count"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Query    string            `json:"query"`
	Messages []MessageListItem `json:"messages"`
}

// printMessageListTable 打印邮件列表表格
func printMessageListTable(messages []MessageListItem) {
	if len(messages) == 0 {
		color.Yellow("No messages found.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Subject", "From", "Date", "Status", "Attach"})
	table.SetBorder(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)

	for _, m := range messages {
		status := "unread"
		if m.IsRead {
			status = "read"
		}

		attach := ""
		if m.HasAttach {
			attach = "yes"
		}

		from := m.FromName
		if from == "" {
			from = m.FromEmail
		}

		idDisplay := m.ID
		if len(idDisplay) > 12 {
			idDisplay = idDisplay[:12]
		}

		row := []string{idDisplay, truncate(m.Subject, 40), from, m.Date, status, attach}
		table.Append(row)
	}

	table.Render()
}

// printMessageDetailTable 打印邮件详情
func printMessageDetailTable(m *MessageDetail) {
	color.Cyan("Subject: %s", m.Subject)

	fromDisplay := m.From.Email
	if m.From.Name != "" {
		fromDisplay = fmt.Sprintf("%s <%s>", m.From.Name, m.From.Email)
	}
	color.Green("From: %s", fromDisplay)

	if len(m.To) > 0 {
		color.Green("To: %s", formatAddresses(m.To))
	}

	if len(m.CC) > 0 {
		color.Green("CC: %s", formatAddresses(m.CC))
	}

	color.Yellow("Date: %s", m.Date)
	fmt.Println()

	if len(m.Attachments) > 0 {
		color.Magenta("Attachments:")
		for _, a := range m.Attachments {
			fmt.Printf("  - %s (%s, %s)\n", a.Filename, a.MimeType, formatSize(a.Size))
		}
		fmt.Println()
	}

	color.White("Body:")
	fmt.Println(m.Body)
}

// printAccountListTable 打印账户列表表格
func printAccountListTable(accounts []AccountListItem) {
	if len(accounts) == 0 {
		color.Yellow("No accounts found.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Email", "Status", "Default"})
	table.SetBorder(false)

	for _, a := range accounts {
		status := "active"
		if !a.IsActive {
			status = "inactive"
		}

		defaultMark := ""
		if a.IsDefault {
			defaultMark = "*"
		}

		table.Append([]string{a.ID, a.Name, a.Email, status, defaultMark})
	}

	table.Render()
}

// printFolderListTable 打印文件夹列表表格
func printFolderListTable(folders []FolderListItem) {
	if len(folders) == 0 {
		color.Yellow("No folders found.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Type", "Messages"})
	table.SetBorder(false)

	for _, f := range folders {
		table.Append([]string{f.ID, f.Name, f.Type, fmt.Sprintf("%d", f.MessageCount)})
	}

	table.Render()
}

// printSearchResultTable 打印搜索结果
func printSearchResultTable(r *SearchResult) {
	color.Cyan("Search query: %s", r.Query)
	fmt.Println()
	printMessageListTable(r.Messages)
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatAddresses 格式化地址列表
func formatAddresses(addrs []AddressInfo) string {
	strs := make([]string, len(addrs))
	for i, a := range addrs {
		if a.Name != "" {
			strs[i] = fmt.Sprintf("%s <%s>", a.Name, a.Email)
		} else {
			strs[i] = a.Email
		}
	}
	return strings.Join(strs, ", ")
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%dKB", size/1024)
	} else {
		return fmt.Sprintf("%dMB", size/1024/1024)
	}
}