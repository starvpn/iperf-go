package main

import (
	"fmt"
	"iperf-go/pkg/iperf"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 简单的持续运行服务器 - 通过循环运行原始服务器实现持续服务
func main() {
	fmt.Println("=== 简单的持续运行服务器 ===\n")
	fmt.Println("服务器监听端口 5201")
	fmt.Println("说明：每处理完一个测试后自动重启以接受下一个连接")
	fmt.Println("按 Ctrl+C 退出\n")

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n接收到退出信号，正在关闭...")
		os.Exit(0)
	}()

	// 设置命令行参数（模拟 -s 参数）
	os.Args = []string{"iperf-go", "-s"}

	// 循环运行服务器
	testCount := 0
	for {
		testCount++
		fmt.Printf("\n========== 测试会话 #%d ==========\n", testCount)
		fmt.Printf("[%s] 等待客户端连接...\n", time.Now().Format("15:04:05"))

		// 创建新的测试实例
		test := iperf.NewIperfTest()
		if test == nil {
			log.Printf("创建测试实例失败")
			time.Sleep(1 * time.Second)
			continue
		}

		// 初始化
		test.Init()

		// 解析参数（这会设置服务器模式）
		if rtn := test.ParseArguments(); rtn < 0 {
			log.Printf("解析参数失败: %v", rtn)
			time.Sleep(1 * time.Second)
			continue
		}

		// 运行测试
		startTime := time.Now()
		rtn := test.RunTest()
		duration := time.Since(startTime)

		if rtn < 0 {
			fmt.Printf("[%s] 测试失败 (错误码: %d)\n", time.Now().Format("15:04:05"), rtn)
			if rtn == -2 {
				// Accept失败，可能是端口被占用
				fmt.Println("提示：可能端口被占用或客户端断开，继续等待...")
			}
			// 短暂等待后重试
			time.Sleep(500 * time.Millisecond)
		} else {
			fmt.Printf("[%s] 测试完成 (耗时: %.2f秒)\n", time.Now().Format("15:04:05"), duration.Seconds())
			fmt.Println("测试结果已显示在上方")

			// 短暂等待以确保资源释放
			time.Sleep(500 * time.Millisecond)
		}

		// 清理资源
		test.FreeTest()

		fmt.Printf("[%s] 准备接受下一个连接...\n", time.Now().Format("15:04:05"))
		fmt.Println("----------------------------------------")
	}
}
