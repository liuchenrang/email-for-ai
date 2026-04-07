package cli

import (
	"context"

	"github.com/chenji/email/internal/container"
	"github.com/chenji/email/pkg/mime"
	"github.com/spf13/cobra"
)

// NewSearchCmd 创建搜索命令
func NewSearchCmd() *cobra.Command {
	var (
		from    string
		to      string
		subject string
		since   string
		until   string
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "搜索邮件",
		Long:  "使用全文搜索查找邮件",
		Example: `  email search "project"                 # 搜索关键词
  email search --from "john@example.com"  # 按发件人搜索
  email search --subject "meeting"        # 按主题搜索
  email search --since "2024-01-01"       # 按时间范围搜索
  email search "urgent" -o json           # JSON格式输出`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]

			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 获取默认账户或指定账户
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

			// 搜索邮件
			messages, err := c.GetMailService().SearchMessages(context.Background(), accID, query, limit)
			if err != nil {
				output := NewErrorOutput("search_error", err.Error())
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

			result := &SearchResult{
				Query:    query,
				Messages: items,
			}

			output := NewSuccessOutput(result)
			output.PrintAndExit()
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "发件人筛选")
	cmd.Flags().StringVar(&to, "to", "", "收件人筛选")
	cmd.Flags().StringVar(&subject, "subject", "", "主题筛选")
	cmd.Flags().StringVar(&since, "since", "", "开始时间")
	cmd.Flags().StringVar(&until, "until", "", "结束时间")
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "结果数量限制")

	return cmd
}