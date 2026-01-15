package cmd // 包名

import (
	"context"          // 上下文管理
	"fmt"              // 格式化字符串
	"net/http"         // HTTP服务
	_ "net/http/pprof" // 性能分析工具pprof
	"os"               // 操作系统接口
	"os/signal"        // 信号处理
	"sync"             // 并发同步
	"syscall"          // 系统调用

	"github.com/spf13/cobra" // 命令行工具库
	"go.uber.org/zap"        // 日志库

	"github.com/yaoxc/EasySwapBase/logger/xzap"    // 自定义日志封装
	"github.com/yaoxc/EasySwapSync/service"        // 服务包
	"github.com/yaoxc/EasySwapSync/service/config" // 配置包
)

// DaemonCmd 定义了 daemon 子命令
var DaemonCmd = &cobra.Command{
	Use:   "daemon",                     // 命令名称
	Short: "sync easy swap order info.", // 简短描述
	Long:  "sync easy swap order info.", // 详细描述
	Run: func(cmd *cobra.Command, args []string) { // 命令执行入口
		fmt.Println("==== daemon Run begin ====")

		wg := &sync.WaitGroup{}                // 创建WaitGroup用于等待goroutine结束
		wg.Add(1)                              // 增加计数
		ctx := context.Background()            // 创建根context
		ctx, cancel := context.WithCancel(ctx) // 创建可取消的context

		fmt.Println("Starting EasySwapSync daemon...")

		onSyncExit := make(chan error, 1) // 服务退出信号通道

		// ==========================================
		// 启动主服务goroutine 【同时也是一个子goroutine】
		// ==========================================
		// 这一行代码开启的是一个子 goroutine。
		// 它通过 go 关键字启动，属于并发执行的独立任务，不会阻塞主流程。
		// 主 goroutine 会等待这个子 goroutine 完成（通过 WaitGroup），
		// 这样可以实现服务的并发处理和优雅退出
		go func() {
			defer wg.Done() // 退出时Done

			cfg, err := config.UnmarshalCmdConfig() // 解析配置文件
			if err != nil {
				xzap.WithContext(ctx).Error("Failed to unmarshal config", zap.Error(err)) // 配置解析失败日志
				onSyncExit <- err                                                         // 通知主流程退出
				return
			}

			// 初始化日志
			_, err = xzap.SetUp(*cfg.Log)
			if err != nil {
				// 日志初始化失败
				xzap.WithContext(ctx).Error("Failed to set up logger", zap.Error(err))
				onSyncExit <- err
				return
			}
			// 启动日志
			xzap.WithContext(ctx).Info("sync server start", zap.Any("config", cfg))

			// 创建服务实例【初始化Redis 、DB、 chainClient】
			s, err := service.New(ctx, cfg)
			if err != nil {
				xzap.WithContext(ctx).Error("Failed to create sync server", zap.Error(err)) // 服务创建失败
				onSyncExit <- err
				return
			}

			// 启动服务
			if err := s.Start(); err != nil {
				xzap.WithContext(ctx).Error("Failed to start sync server", zap.Error(err)) // 启动失败
				onSyncExit <- err
				return
			}

			// 如果配置项 cfg.Monitor.PprofEnable 为 true，
			// 则启动 pprof 性能分析服务。
			// 具体来说，http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Monitor.PprofPort), nil) 会在指定端口启动一个 HTTP 服务器，
			// 暴露 Go 的运行时性能分析接口（如 /debug/pprof），方便开发者通过浏览器或工具远程分析程序的 CPU、内存等性能数据，用于排查和优化性能瓶颈。
			if cfg.Monitor.PprofEnable {
				fmt.Println("Starting pprof server on port：", cfg.Monitor.PprofPort)
				http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Monitor.PprofPort), nil) // 启动pprof服务
			}
		}()

		// 系统信号通道
		onSignal := make(chan os.Signal)
		// 监听中断和终止信号
		signal.Notify(onSignal, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-onSignal: // 收到信号
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
				cancel()                                                                         // 取消context，优雅退出
				xzap.WithContext(ctx).Info("Exit by signal", zap.String("signal", sig.String())) // 记录信号退出日志
			}
		case err := <-onSyncExit: // 服务异常退出
			cancel()                                                     // 取消context
			xzap.WithContext(ctx).Error("Exit by error", zap.Error(err)) // 记录错误日志
		}
		wg.Wait() // 等待所有goroutine结束

		fmt.Println("==== daemon Run end ====")

	},
}

// init 函数将 DaemonCmd 注册到主命令
func init() {
	fmt.Println("执行到daemon类中")
	rootCmd.AddCommand(DaemonCmd) // 注册子命令
}
