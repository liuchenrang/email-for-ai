package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	output    string
	accountID string
	quietFlag bool
)

// Execute 执行CLI入口
func Execute(version string) {
	rootCmd := NewRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// NewRootCmd 创建根命令
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email",
		Short: "CLI邮件客户端",
		Long: `一个基于DDD架构的命令行邮件客户端。

支持IMAP/POP3收件、SMTP发件，本地SQLite存储和全文搜索。
所有命令支持JSON输出格式（--output json），方便AI工具调用。`,
		Version: version,
	}

	// 全局标志
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")
	cmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "输出格式: table|json")
	cmd.PersistentFlags().StringVarP(&accountID, "account", "a", "", "指定账户ID")
	cmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "静默模式，仅输出错误")

	// 绑定到 viper
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("output", cmd.PersistentFlags().Lookup("output"))

	// 初始化配置
	cobra.OnInitialize(initConfig)

	// 添加子命令
	cmd.AddCommand(NewAccountCmd())
	cmd.AddCommand(NewSyncCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewReadCmd())
	cmd.AddCommand(NewSearchCmd())
	cmd.AddCommand(NewComposeCmd())
	cmd.AddCommand(NewSendCmd())
	cmd.AddCommand(NewFoldersCmd())
	cmd.AddCommand(NewConfigCmd())

	return cmd
}

// initConfig 初始化配置
func initConfig() {
	// 设置静默模式
	SetQuietMode(quietFlag)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// 默认配置路径
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		// 搜索配置
		viper.AddConfigPath(home + "/.email")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// 环境变量
	viper.SetEnvPrefix("EMAIL")
	viper.AutomaticEnv()

	// 读取配置
	if err := viper.ReadInConfig(); err == nil {
		// 配置文件存在
	} else {
		// 配置文件不存在时使用默认值
		initDefaultConfig()
	}
}

// initDefaultConfig 初始化默认配置
func initDefaultConfig() {
	home, _ := os.UserHomeDir()

	viper.SetDefault("database.path", home+"/.email/data.db")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("sync.protocol", "imap")
	viper.SetDefault("sync.limit", 100)
	viper.SetDefault("display.format", "table")
	viper.SetDefault("display.date_format", "2006-01-02 15:04")
	viper.SetDefault("display.max_body_length", 500)
}

// getOutputFormat 获取输出格式
func getOutputFormat() string {
	if output != "" {
		return output
	}
	return viper.GetString("display.format")
}

// isJSONOutput 是否为JSON输出
func isJSONOutput() bool {
	return getOutputFormat() == "json"
}

// getAccountID 获取账户ID
func getAccountID() string {
	return accountID
}