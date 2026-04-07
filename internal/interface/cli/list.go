package cli

import (
	"context"

	"github.com/chenji/email/internal/container"
	"github.com/chenji/email/pkg/mime"
	"github.com/spf13/cobra"
)

// NewListCmd 创建列表命令
func NewListCmd() *cobra.Command {
	var (
		folder string
		unread bool
		limit  int
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出邮件",
		Long:    "列出本地存储的邮件",
		Aliases: []string{"ls"},
		Example: `  email list                      # 列出INBOX邮件
  email list --folder Sent        # 列出已发送邮件
  email list --unread             # 仅显示未读邮件
  email list -n 10                # 限制显示数量
  email list -o json              # JSON格式输出`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 获取默认账户
			accountIDVal := getAccountID()
			var accID string
			if accountIDVal != "" {
				accID = accountIDVal
			} else {
				acc, err := c.GetAccountService().GetDefaultAccount(context.Background())
				if err != nil {
					output := NewErrorOutput("account_error", "请先添加账户")
					output.PrintAndExit()
					return
				}
				accID = acc.ID().String()
			}

			// 获取邮件列表
			messages, err := c.GetMailService().ListMessages(context.Background(), accID, folder, unread, limit)
			if err != nil {
				output := NewErrorOutput("list_error", err.Error())
				output.PrintAndExit()
				return
			}

			// 转换为输出格式
			items := make([]MessageListItem, len(messages))
			for i, msg := range messages {
				// 解码邮件头
				subject := mime.DecodeSubject(msg.Subject())
				fromName := mime.DecodeAddress(msg.From().Name())
				fromEmail := msg.From().Email()
				date := msg.Date().Format("2006-01-02 15:04")

				items[i] = MessageListItem{
					ID:        msg.ID().String(),
					Subject:   subject,
					FromName:  fromName,
					FromEmail: fromEmail,
					Date:      date,
					IsRead:    msg.IsRead(),
					HasAttach: msg.HasAttachments(),
				}
			}

			output := NewSuccessOutput(items)
			output.PrintAndExit()
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "INBOX", "文件夹名称")
	cmd.Flags().BoolVarP(&unread, "unread", "u", false, "仅显示未读邮件")
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "显示数量限制")

	return cmd
}