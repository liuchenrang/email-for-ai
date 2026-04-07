package cli

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/chenji/email/internal/application/service"
	"github.com/chenji/email/internal/container"
	"github.com/chenji/email/internal/domain/email"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewSyncCmd 创建同步命令
func NewSyncCmd() *cobra.Command {
	var (
		protocol string
		folder   string
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步邮件",
		Long:  "从邮件服务器同步邮件到本地数据库，支持钩子通知",
		Example: `  email sync                      # 使用默认IMAP协议同步
  email sync --protocol pop3     # 使用POP3协议同步
  email sync --folder INBOX      # 仅同步指定文件夹
  email sync --limit 50          # 限制同步数量`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 获取账户
			accountIDVal := getAccountID()
			var accountID string
			if accountIDVal != "" {
				accountID = accountIDVal
			} else {
				acc, err := c.GetAccountService().GetDefaultAccount(context.Background())
				if err != nil {
					output := NewErrorOutput("account_error", "请先添加账户")
					output.PrintAndExit()
					return
				}
				accountID = acc.ID().String()
			}

			// 同步邮件
			result, err := c.GetSyncService().Sync(context.Background(), accountID, protocol, limit)
			if err != nil {
				output := NewErrorOutput("sync_error", err.Error())
				output.PrintAndExit()
				return
			}

			// 执行新邮件钩子
			hookCmd := viper.GetString("hooks.new_mail")
			hookEnabled := viper.GetBool("hooks.enabled")
			if hookEnabled && hookCmd != "" && result.New > 0 {
				// 获取新邮件并触发钩子
				messages, _ := c.GetMailService().ListMessages(context.Background(), accountID, "", true, result.New)
				for _, msg := range messages {
					// 通过JSON传递给钩子
					triggerNewMailHook(msg, hookCmd)
				}
			}

			output := NewSuccessOutput(map[string]interface{}{
				"account_id": accountID,
				"protocol":   protocol,
				"synced":     result.Synced,
				"new":        result.New,
				"skipped":    result.Skipped,
				"folder":     folder,
			})
			output.PrintAndExit()
		},
	}

	cmd.Flags().StringVar(&protocol, "protocol", "imap", "协议类型: imap|pop3")
	cmd.Flags().StringVar(&folder, "folder", "", "指定文件夹")
	cmd.Flags().IntVar(&limit, "limit", 100, "同步数量限制")

	return cmd
}

// triggerNewMailHook 触发新邮件钩子
func triggerNewMailHook(msg *email.Message, hookCmd string) {
	toStrs := make([]string, len(msg.To()))
	for i, addr := range msg.To() {
		toStrs[i] = addr.Email()
	}

	bodyPreview := msg.Body().Text()
	if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500]
	}

	payload := service.HookPayload{
		Event:     "new_mail",
		Timestamp: time.Now().Format(time.RFC3339),
		AccountID: msg.AccountID().String(),
		Message: service.MessageInfo{
			ID:          msg.ID().String(),
			MessageID:   msg.MessageID(),
			Subject:     msg.Subject(),
			FromName:    msg.From().Name(),
			FromEmail:   msg.From().Email(),
			To:          strings.Join(toStrs, ", "),
			Date:        msg.Date().Format(time.RFC3339),
			BodyPreview: bodyPreview,
			IsRead:      msg.IsRead(),
		},
	}

	payloadJSON, _ := json.Marshal(payload)

	// 执行钩子
	execCmd := exec.Command("sh", "-c", hookCmd)
	execCmd.Stdin = strings.NewReader(string(payloadJSON))
	execCmd.Run() // 忽略错误，不阻塞同步
}