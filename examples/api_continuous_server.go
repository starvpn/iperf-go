package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iperf-go/pkg/iperf"
)

// 这个示例展示如何使用API库创建持续运行的服务器
func main() {
	fmt.Println("=== API库方式 - 持续运行服务器示例 ===\n")

	// 创建服务器配置
	config := iperf.ServerConfig(5208) // 使用5208端口
	config.Duration = 10 * time.Second
	config.Interval = 1 * time.Second
	config.Protocol = "tcp"

	// 创建持续运行服务器
	server, err := iperf.NewContinuousServer(config)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 统计信息
	var totalTests int
	var totalBytes uint64
	var totalDuration time.Duration

	// 设置事件处理器
	server.SetEventHandler(func(event iperf.Event) {
		switch event.Type {
		case iperf.EventConnected:
			if data, ok := event.Data.(map[string]interface{}); ok {
				if testNum, ok := data["test_num"].(int); ok {
					fmt.Printf("\n📡 [事件] 客户端连接 (测试 #%d)\n", testNum)
				}
			}

		case iperf.EventInterval:
			// 间隔报告（如果需要可以处理）

		case iperf.EventComplete:
			if result, ok := event.Data.(*iperf.TestResult); ok {
				totalTests++
				totalBytes += result.TotalBytes
				totalDuration += result.Duration

				fmt.Printf("✅ [事件] 测试完成:\n")
				fmt.Printf("   - 传输数据: %.2f MB\n", float64(result.TotalBytes)/1024/1024)
				fmt.Printf("   - 带宽: %.2f Mbps\n", result.Bandwidth)
				fmt.Printf("   - 耗时: %.2f 秒\n", result.Duration.Seconds())

				// 显示累计统计
				if totalTests > 1 {
					avgBandwidth := float64(totalBytes*8) / totalDuration.Seconds() / 1000000
					fmt.Printf("\n📊 累计统计:\n")
					fmt.Printf("   - 总测试数: %d\n", totalTests)
					fmt.Printf("   - 总传输: %.2f MB\n", float64(totalBytes)/1024/1024)
					fmt.Printf("   - 平均带宽: %.2f Mbps\n", avgBandwidth)
				}
			}

		case iperf.EventError:
			fmt.Printf("❌ [事件] 错误: %v\n", event.Error)
		}
	})

	// 启动服务器
	fmt.Printf("启动服务器，监听端口 %d...\n", config.Port)
	if err := server.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	fmt.Println("服务器正在运行...")
	fmt.Println("使用以下命令测试:")
	fmt.Printf("  ./iperf-go -c localhost -p %d\n", config.Port)
	fmt.Println("\n按 Ctrl+C 退出")
	fmt.Println("========================================")

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 定期显示状态
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := server.GetTestCount()
			if count > 0 {
				fmt.Printf("\n⏱️  [状态] 服务器运行中... 已处理 %d 个测试\n", count)
			}

		case <-sigChan:
			fmt.Println("\n\n接收到退出信号...")
			server.Stop()

			// 显示最终统计
			fmt.Println("\n========================================")
			fmt.Println("📈 最终统计:")
			fmt.Printf("   总测试数: %d\n", totalTests)
			if totalTests > 0 {
				fmt.Printf("   总传输: %.2f MB\n", float64(totalBytes)/1024/1024)
				if totalDuration > 0 {
					avgBandwidth := float64(totalBytes*8) / totalDuration.Seconds() / 1000000
					fmt.Printf("   平均带宽: %.2f Mbps\n", avgBandwidth)
				}
			}
			fmt.Println("========================================")
			fmt.Println("\n服务器已关闭")
			return
		}
	}
}
