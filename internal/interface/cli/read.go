package cli

import (
	"context"
	"strings"

	"github.com/chenji/email/internal/container"
	"github.com/chenji/email/pkg/mime"
	"github.com/spf13/cobra"
)

// NewReadCmd 创建阅读命令
func NewReadCmd() *cobra.Command {
	var raw bool

	cmd := &cobra.Command{
		Use:   "read <message-id>",
		Short: "阅读邮件",
		Long:  "显示邮件详细内容",
		Example: `  email read msg_abc123        # 显示邮件内容
  email read msg_abc123 --raw  # 显示原始邮件内容`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			messageID := args[0]

			// 初始化数据库
			c := container.GetContainer()
			if err := c.InitDB(); err != nil {
				output := NewErrorOutput("db_error", err.Error())
				output.PrintAndExit()
				return
			}
			defer c.Close()

			// 获取邮件详情
			msg, err := c.GetMailService().GetMessage(context.Background(), messageID)
			if err != nil {
				output := NewErrorOutput("read_error", err.Error())
				output.PrintAndExit()
				return
			}

			// 解码邮件头
			subject := mime.DecodeSubject(msg.Subject())
			fromName := mime.DecodeAddress(msg.From().Name())
			fromEmail := msg.From().Email()

			// 构建收件人列表
			toList := make([]AddressInfo, len(msg.To()))
			for i, addr := range msg.To() {
				toList[i] = AddressInfo{
					Name:  mime.DecodeAddress(addr.Name()),
					Email: addr.Email(),
				}
			}

			// 构建抄送列表
			ccList := make([]AddressInfo, len(msg.CC()))
			for i, addr := range msg.CC() {
				ccList[i] = AddressInfo{
					Name:  mime.DecodeAddress(addr.Name()),
					Email: addr.Email(),
				}
			}

			// 获取正文（存储的是原始MIME内容，需要解析）
			rawBody := msg.Body().Text()
			bodyContent := rawBody

			// 尝试解析MIME multipart内容
			if rawBody != "" && (strings.Contains(rawBody, "boundary=") || strings.Contains(rawBody, "Content-Type:")) {
				bodyContent = mime.ExtractReadableBody(rawBody)
			}

			// 如果正文为空，尝试HTML
			if bodyContent == "" && msg.Body().HTML() != "" {
				bodyContent = mime.ExtractReadableBody(msg.Body().HTML())
			}

			// 构建详情输出
			detail := &MessageDetail{
				ID:        msg.ID().String(),
				MessageID: msg.MessageID(),
				Subject:   subject,
				From:      AddressInfo{Name: fromName, Email: fromEmail},
				To:        toList,
				CC:        ccList,
				Date:      msg.Date().Format("2006-01-02 15:04:05"),
				Body:      bodyContent,
				IsRead:    msg.IsRead(),
			}

			// 转换Flags为字符串数组
			flags := msg.Flags()
			flagStrs := make([]string, len(flags))
			for i, f := range flags {
				flagStrs[i] = string(f)
			}
			detail.Flags = flagStrs

			// 附件列表
			if len(msg.Attachments()) > 0 {
				detail.Attachments = make([]AttachmentInfo, len(msg.Attachments()))
				for i, att := range msg.Attachments() {
					detail.Attachments[i] = AttachmentInfo{
						Filename: att.Filename(),
						MimeType: att.MimeType(),
						Size:     att.Size(),
					}
				}
			}

			output := NewSuccessOutput(detail)
			output.PrintAndExit()
		},
	}

	cmd.Flags().BoolVar(&raw, "raw", false, "显示原始邮件内容")

	return cmd
}