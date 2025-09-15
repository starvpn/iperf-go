package main

import (
	"fmt"
	"iperf-go/pkg/iperf"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 直接设置参数，避免flag重复定义
	// 从命令行获取端口（如果有的话）
	port := uint(5201)
	showHelp := false
	debug := false
	info := false

	// 简单解析参数
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-h", "--help":
			showHelp = true
		case "-p":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &port)
				i++
			}
		case "-debug":
			debug = true
		case "-info":
			info = true
		}
	}

	if showHelp {
		fmt.Println("iperf-go 持续运行服务器")
		fmt.Println("\n用法:")
		fmt.Println("  iperf-server-loop [选项]")
		fmt.Println("\n选项:")
		fmt.Println("  -h, --help    显示帮助")
		fmt.Println("  -p PORT       监听端口 (默认: 5201)")
		fmt.Println("  -debug        调试模式")
		fmt.Println("  -info         信息模式")
		fmt.Println("\n特性:")
		fmt.Println("  - 服务器会持续运行，自动处理多个客户端连接")
		fmt.Println("  - 每个测试完成后自动重置并等待下一个连接")
		fmt.Println("  - 支持所有标准iperf协议")
		fmt.Println("\n示例:")
		fmt.Println("  iperf-server-loop -p 5201")
		fmt.Println("  iperf-server-loop -p 5201 -debug")
		os.Exit(0)
	}

	fmt.Println("====================================")
	fmt.Println("    iperf-go 持续运行服务器")
	fmt.Println("====================================")
	fmt.Printf("监听端口: %d\n", port)
	fmt.Println("服务器将持续接受客户端连接")
	fmt.Println("按 Ctrl+C 优雅退出")
	fmt.Println("------------------------------------\n")

	// 创建测试实例
	test := iperf.NewIperfTest()
	if test == nil {
		fmt.Println("错误：创建测试实例失败")
		os.Exit(1)
	}

	// 初始化
	test.Init()

	// 构建参数
	args := []string{"iperf-server-loop"}

	// 添加服务器标志
	args = append(args, "-s")

	// 添加端口
	args = append(args, "-p", fmt.Sprintf("%d", port))

	// 添加日志级别
	if debug {
		args = append(args, "-debug")
	} else if info {
		args = append(args, "-info")
	}

	// 设置参数并解析
	os.Args = args
	if rtn := test.ParseArguments(); rtn < 0 {
		fmt.Printf("错误：解析参数失败: %v\n", rtn)
		os.Exit(1)
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n接收到退出信号...")
		fmt.Println("正在优雅关闭服务器...")
		// 清理资源
		test.FreeTest()
		fmt.Println("服务器已关闭")
		os.Exit(0)
	}()

	// 运行持续服务器
	if rtn := test.RunServerLoop(); rtn < 0 {
		fmt.Printf("错误：服务器运行失败: %v\n", rtn)
		os.Exit(1)
	}
}
