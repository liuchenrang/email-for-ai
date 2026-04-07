package cli

import (
	"github.com/spf13/cobra"
)

// NewFoldersCmd 创建文件夹命令
func NewFoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "文件夹管理",
		Long:  "管理邮件文件夹",
	}

	cmd.AddCommand(NewFoldersListCmd())
	cmd.AddCommand(NewFoldersCreateCmd())
	cmd.AddCommand(NewFoldersDeleteCmd())

	return cmd
}

// NewFoldersListCmd 创建列出文件夹命令
func NewFoldersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出文件夹",
		Long:    "列出所有邮件文件夹",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 实现列出文件夹逻辑
			folders := []FolderListItem{
				{ID: "fld_inbox", Name: "INBOX", Type: "inbox", MessageCount: 42},
				{ID: "fld_sent", Name: "Sent", Type: "sent", MessageCount: 15},
				{ID: "fld_drafts", Name: "Drafts", Type: "drafts", MessageCount: 3},
				{ID: "fld_trash", Name: "Trash", Type: "trash", MessageCount: 8},
				{ID: "fld_custom", Name: "Work", Type: "custom", MessageCount: 20},
			}
			output := NewSuccessOutput(folders)
			output.PrintAndExit()
		},
	}

	return cmd
}

// NewFoldersCreateCmd 创建创建文件夹命令
func NewFoldersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <folder-name>",
		Short: "创建文件夹",
		Long:  "创建新的邮件文件夹",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 实现创建文件夹逻辑
			output := NewSuccessOutput(map[string]string{
				"id":      "fld_new",
				"name":    args[0],
				"message": "Folder created successfully",
			})
			output.PrintAndExit()
		},
	}

	return cmd
}

// NewFoldersDeleteCmd 创建删除文件夹命令
func NewFoldersDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <folder-id>",
		Short:   "删除文件夹",
		Long:    "删除邮件文件夹（仅限自定义文件夹）",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 实现删除文件夹逻辑
			output := NewSuccessOutput(map[string]string{
				"message": "Folder deleted successfully",
			})
			output.PrintAndExit()
		},
	}

	return cmd
}