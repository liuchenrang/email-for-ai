package cli

import (
	"context"

	"github.com/chenji/email/internal/application/service"
	"github.com/chenji/email/internal/container"
	"github.com/spf13/cobra"
)

// NewAccountCmd 创建账户管理命令
func NewAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "账户管理",
		Long:  "管理邮件账户配置",
	}

	cmd.AddCommand(NewAccountAddCmd())
	cmd.AddCommand(NewAccountListCmd())
	cmd.AddCommand(NewAccountRemoveCmd())
	cmd.AddCommand(NewAccountSetDefaultCmd())

	return cmd
}

// NewAccountAddCmd 创建添加账户命令
func NewAccountAddCmd() *cobra.Command {
	var (
		name      string
		email     string
		imapHost  string
		imapPort  int
		imapUser  string
		imapPass  string
		smtpHost  string
		smtpPort  int
		smtpUser  string
		smtpPass  string
		pop3Host  string
		pop3Port  int
		pop3User  string
		pop3Pass  string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "添加邮件账户",
		Long:  "添加新的邮件账户配置",
		Example: `  email account add \
    --name "Gmail" \
    --email "user@gmail.com" \
    --imap-host "imap.gmail.com" \
    --imap-port 993 \
    --imap-pass "your-app-password" \
    --smtp-host "smtp.gmail.com" \
    --smtp-port 587 \
    --smtp-pass "your-app-password"`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 设置默认用户名（如果未指定则使用邮箱）
			if imapUser == "" {
				imapUser = email
			}
			if smtpUser == "" {
				smtpUser = email
			}

			// 获取服务
			accountService := c.GetAccountService()

			// 添加账户
			acc, err := accountService.AddAccount(context.Background(), service.AddAccountInput{
				Name:     name,
				Email:    email,
				IMAPHost: imapHost,
				IMAPPort: imapPort,
				IMAPUser: imapUser,
				IMAPPass: imapPass,
				SMTPHost: smtpHost,
				SMTPPort: smtpPort,
				SMTPUser: smtpUser,
				SMTPPass: smtpPass,
				POP3Host: pop3Host,
				POP3Port: pop3Port,
				POP3User: pop3User,
				POP3Pass: pop3Pass,
			})

			if err != nil {
				output := NewErrorOutput("add_account_error", err.Error())
				output.PrintAndExit()
				return
			}

			output := NewSuccessOutput(map[string]string{
				"id":      acc.ID().String(),
				"name":    acc.Name(),
				"email":   acc.Email(),
				"message": "Account added successfully. Run 'email sync' to fetch messages.",
			})
			output.PrintAndExit()
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "账户名称")
	cmd.Flags().StringVar(&email, "email", "", "邮箱地址")
	cmd.Flags().StringVar(&imapHost, "imap-host", "", "IMAP服务器地址")
	cmd.Flags().IntVar(&imapPort, "imap-port", 993, "IMAP服务器端口")
	cmd.Flags().StringVar(&imapUser, "imap-user", "", "IMAP用户名 (默认使用邮箱)")
	cmd.Flags().StringVar(&imapPass, "imap-pass", "", "IMAP密码")
	cmd.Flags().StringVar(&smtpHost, "smtp-host", "", "SMTP服务器地址")
	cmd.Flags().IntVar(&smtpPort, "smtp-port", 587, "SMTP服务器端口")
	cmd.Flags().StringVar(&smtpUser, "smtp-user", "", "SMTP用户名 (默认使用邮箱)")
	cmd.Flags().StringVar(&smtpPass, "smtp-pass", "", "SMTP密码")
	cmd.Flags().StringVar(&pop3Host, "pop3-host", "", "POP3服务器地址 (可选)")
	cmd.Flags().IntVar(&pop3Port, "pop3-port", 110, "POP3服务器端口")
	cmd.Flags().StringVar(&pop3User, "pop3-user", "", "POP3用户名")
	cmd.Flags().StringVar(&pop3Pass, "pop3-pass", "", "POP3密码")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("imap-host")
	cmd.MarkFlagRequired("smtp-host")

	return cmd
}

// NewAccountListCmd 创建列出账户命令
func NewAccountListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出所有账户",
		Long:    "列出所有已配置的邮件账户",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 获取账户列表
			accounts, err := c.GetAccountService().ListAccounts(context.Background())
			if err != nil {
				output := NewErrorOutput("list_error", err.Error())
				output.PrintAndExit()
				return
			}

			// 转换为输出格式
			items := make([]AccountListItem, len(accounts))
			for i, acc := range accounts {
				items[i] = AccountListItem{
					ID:        acc.ID().String(),
					Name:      acc.Name(),
					Email:     acc.Email(),
					IsActive:  acc.IsActive(),
					IsDefault: acc.IsDefault(),
				}
			}

			output := NewSuccessOutput(items)
			output.PrintAndExit()
		},
	}

	return cmd
}

// NewAccountRemoveCmd 创建删除账户命令
func NewAccountRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <account-id>",
		Short:   "删除账户",
		Long:    "删除指定的邮件账户",
		Aliases: []string{"rm", "delete"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			accountID := args[0]

			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 删除账户
			if err := c.GetAccountService().RemoveAccount(context.Background(), accountID); err != nil {
				output := NewErrorOutput("remove_error", err.Error())
				output.PrintAndExit()
				return
			}

			output := NewSuccessOutput(map[string]string{
				"message": "Account removed successfully",
			})
			output.PrintAndExit()
		},
	}

	return cmd
}

// NewAccountSetDefaultCmd 创建设置默认账户命令
func NewAccountSetDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default <account-id>",
		Short: "设置默认账户",
		Long:  "设置指定的账户为默认账户",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			accountID := args[0]

			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 设置默认账户
			if err := c.GetAccountService().SetDefaultAccount(context.Background(), accountID); err != nil {
				output := NewErrorOutput("set_default_error", err.Error())
				output.PrintAndExit()
				return
			}

			output := NewSuccessOutput(map[string]string{
				"message": "Default account updated",
			})
			output.PrintAndExit()
		},
	}

	return cmd
}