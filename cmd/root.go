package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sync",
	Short: "root server.",
	Long:  `root server.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("cfgFile=", cfgFile)
}

func init() {
	// 注册 initConfig 函数，使其在 rootCmd.Execute() 执行前自动调用。用于初始化配置文件和环境变量。
	cobra.OnInitialize(initConfig)
	// 获取 rootCmd 的全局（持久）命令行参数集合，便于后续添加参数
	flags := rootCmd.PersistentFlags()
	// 定义一个名为 config（简写 -c）的字符串参数，
	// 默认值为 ./config/config_import.toml，用于指定配置文件路径。
	// 用户可通过命令行传递该参数，值会存入 cfgFile 变量。
	flags.StringVarP(&cfgFile, "config", "c", "./config/config_import.toml", "config file (default is $HOME/.config_import.toml)")
}

// 在 rootCmd.Execute() 执行前调用
func initConfig() {
	// 判断是否通过命令行参数指定了配置文件路径
	if cfgFile != "" {
		// 如果指定了配置文件路径，设置 viper 直接读取该文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 如果未指定配置文件，获取当前用户的主目录路径 /Users/$HOME$
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 设置 viper 在主目录下查找配置文件。
		viper.AddConfigPath(home)
		// 设置配置文件名为 config_import（不含扩展名）
		viper.SetConfigName("config_import")
	}
	// 允许 viper 自动读取环境变量
	viper.AutomaticEnv()
	// 指定配置文件类型为 TOML
	viper.SetConfigType("toml")
	// 设置环境变量前缀为 EASYSWAP_
	viper.SetEnvPrefix("EasySwap")
	// 创建一个字符串替换器，将点号替换为下划线
	replacer := strings.NewReplacer(".", "_")
	// 设置环境变量名的替换规则（如 db.host → DB_HOST）。
	viper.SetEnvKeyReplacer(replacer)
	// 读取配置文件，成功则打印使用的配置文件路径，失败则抛出异常终止程序
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		panic(err)
	}

}
