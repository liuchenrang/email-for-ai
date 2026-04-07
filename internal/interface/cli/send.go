package cli

import (
	"context"
	"time"

	"github.com/chenji/email/internal/container"
	"github.com/chenji/email/internal/domain/account"
	"github.com/chenji/email/internal/domain/email"
	"github.com/chenji/email/internal/infrastructure/mail/smtp"
	"github.com/spf13/cobra"
)

// NewSendCmd 创建发送命令
func NewSendCmd() *cobra.Command {
	var (
		to      string
		cc      string
		subject string
		body    string
	)

	cmd := &cobra.Command{
		Use:   "send",
		Short: "发送邮件",
		Long:  "发送撰写好的邮件",
		Example: `  email send --to "recipient@example.com" --subject "Hello" --body "Content"
  email compose -t "recipient@example.com" -s "Hello" -b "Body" | email send`,
		Run: func(cmd *cobra.Command, args []string) {
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
			var acc *account.Account
			var err error

			if accountIDVal != "" {
				acc, err = c.GetAccountService().GetAccount(context.Background(), accountIDVal)
			} else {
				acc, err = c.GetAccountService().GetDefaultAccount(context.Background())
			}

			if err != nil {
				output := NewErrorOutput("account_error", "请先添加账户或指定发送账户: "+err.Error())
				output.PrintAndExit()
				return
			}

			// 创建邮件对象
			from := email.NewAddress(acc.Name(), acc.Email())
			toAddrs := []email.Address{email.NewAddress("", to)}

			var ccAddrs []email.Address
			if cc != "" {
				ccAddrs = []email.Address{email.NewAddress("", cc)}
			}

			emailBody := email.NewTextBody(body)

			msg := email.NewMessage(
				acc.ID(),
				email.FolderID(""),
				"",
				subject,
				from,
				toAddrs,
				time.Now(),
				emailBody,
			)

			// 设置抄送
			if len(ccAddrs) > 0 {
				msg.SetCC(ccAddrs)
			}

			// 发送邮件
			smtpClient := smtp.NewSMTPClient(acc.SMTPConfig())
			if err := smtpClient.Send(msg); err != nil {
				output := NewErrorOutput("send_error", err.Error())
				output.PrintAndExit()
				return
			}

			output := NewSuccessOutput(map[string]string{
				"message": "邮件发送成功",
				"from":    acc.Email(),
				"to":      to,
				"subject": subject,
				"sent_at": time.Now().Format("2006-01-02 15:04:05"),
			})
			output.PrintAndExit()
		},
	}

	cmd.Flags().StringVarP(&to, "to", "t", "", "收件人")
	cmd.Flags().StringVar(&cc, "cc", "", "抄送")
	cmd.Flags().StringVarP(&subject, "subject", "s", "", "主题")
	cmd.Flags().StringVarP(&body, "body", "b", "", "正文")

	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("subject")

	return cmd
}