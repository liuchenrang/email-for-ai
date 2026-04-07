package cli

import (
	"github.com/spf13/cobra"
)

// NewComposeCmd 创建撰写命令
func NewComposeCmd() *cobra.Command {
	var (
		to      string
		cc      string
		subject string
		body    string
		attach  string
	)

	cmd := &cobra.Command{
		Use:   "compose",
		Short: "撰写邮件",
		Long:  "撰写新邮件，可通过管道发送",
		Example: `  email compose --to "recipient@example.com" --subject "Hello" --body "Content"
  email compose -t "user@example.com" -s "Test" -b "Body" | email send`,
		Run: func(cmd *cobra.Command, args []string) {
			// 输出邮件信息，供send命令使用
			output := NewSuccessOutput(map[string]string{
				"to":      to,
				"cc":      cc,
				"subject": subject,
				"body":    body,
				"attach":  attach,
			})
			output.PrintAndExit()
		},
	}

	cmd.Flags().StringVarP(&to, "to", "t", "", "收件人")
	cmd.Flags().StringVar(&cc, "cc", "", "抄送")
	cmd.Flags().StringVarP(&subject, "subject", "s", "", "主题")
	cmd.Flags().StringVarP(&body, "body", "b", "", "正文")
	cmd.Flags().StringVar(&attach, "attach", "", "附件路径")

	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("subject")

	return cmd
}