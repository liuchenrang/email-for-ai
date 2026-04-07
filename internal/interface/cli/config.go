package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewConfigCmd 创建配置命令
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
		Long:  "查看和管理应用配置",
	}

	cmd.AddCommand(NewConfigShowCmd())
	cmd.AddCommand(NewConfigInitCmd())
	cmd.AddCommand(NewConfigSetCmd())

	return cmd
}

// NewConfigShowCmd 创建显示配置命令
func NewConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "显示当前配置",
		Long:  "显示当前应用的完整配置",
		Run: func(cmd *cobra.Command, args []string) {
			settings := map[string]interface{}{
				"database": map[string]string{
					"path": viper.GetString("database.path"),
				},
				"log": map[string]string{
					"level": viper.GetString("log.level"),
				},
				"sync": map[string]interface{}{
					"protocol": viper.GetString("sync.protocol"),
					"limit":    viper.GetInt("sync.limit"),
				},
				"display": map[string]interface{}{
					"format":         viper.GetString("display.format"),
					"date_format":    viper.GetString("display.date_format"),
					"max_body_length": viper.GetInt("display.max_body_length"),
				},
			}

			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				configFile = "default (no config file found)"
			}
			settings["config_file"] = configFile

			output := NewSuccessOutput(settings)
			output.PrintAndExit()
		},
	}

	return cmd
}

// NewConfigInitCmd 创建初始化配置命令
func NewConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化配置文件",
		Long:  "创建默认配置文件",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := os.UserHomeDir()
			if err != nil {
				output := NewErrorOutput("home_dir_error", err.Error())
				output.Print()
				return
			}

			configPath := home + "/.email/config.yaml"

			// 检查是否已存在
			if _, err := os.Stat(configPath); err == nil {
				output := NewErrorOutput("config_exists", "Config file already exists at " + configPath)
				output.Print()
				return
			}

			// 创建目录
			if err := os.MkdirAll(home+"/.email", 0755); err != nil {
				output := NewErrorOutput("mkdir_error", err.Error())
				output.Print()
				return
			}

			// 写入默认配置
			defaultConfig := `# Email CLI Client Configuration
database:
  path: ~/.email/data.db

log:
  level: info

sync:
  protocol: imap
  limit: 100

display:
  format: table
  date_format: "2006-01-02 15:04"
  max_body_length: 500
`

			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				output := NewErrorOutput("write_error", err.Error())
				output.Print()
				return
			}

			output := NewSuccessOutput(map[string]string{
				"message":    "Config file created successfully",
				"config_path": configPath,
			})
			output.PrintAndExit()
		},
	}

	return cmd
}

// NewConfigSetCmd 创建设置配置命令
func NewConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "设置配置项",
		Long:  "设置指定的配置项值",
		Example: `  email config set display.format json
  email config set sync.limit 200`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			viper.Set(key, value)

			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				output := NewErrorOutput("no_config_file", "No config file found. Run 'email config init' first.")
				output.Print()
				return
			}

			if err := viper.WriteConfig(); err != nil {
				output := NewErrorOutput("write_error", err.Error())
				output.Print()
				return
			}

			output := NewSuccessOutput(map[string]string{
				"message": fmt.Sprintf("Config updated: %s = %s", key, value),
			})
			output.PrintAndExit()
		},
	}

	return cmd
}