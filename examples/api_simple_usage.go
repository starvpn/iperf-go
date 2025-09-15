package main

import (
	"fmt"
	"log"
	"time"

	"iperf-go/pkg/iperf"
)

// 简单的API使用示例 - 展示最基础的用法
func main() {
	fmt.Println("=== 简单API使用示例 ===\n")

	// 方式1: 使用ContinuousServer（持续运行）
	useContinuousServer()

	// 方式2: 使用原有Server API的StartContinuous方法
	// useOriginalServerAPI()
}

// useContinuousServer 使用新的ContinuousServer API
func useContinuousServer() {
	fmt.Println("📦 使用 ContinuousServer API\n")

	// 1. 创建配置
	config := &iperf.Config{
		Role:     iperf.RoleServer,
		Port:     5209,
		Protocol: "tcp",
		Duration: 5 * time.Second,
		Interval: 1 * time.Second,
	}

	// 2. 创建服务器
	server, err := iperf.NewContinuousServer(config)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 3. 设置简单的事件处理器（可选）
	server.SetEventHandler(func(event iperf.Event) {
		switch event.Type {
		case iperf.EventComplete:
			fmt.Println("✅ 一个测试完成")
		case iperf.EventError:
			fmt.Printf("❌ 错误: %v\n", event.Error)
		}
	})

	// 4. 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}

	fmt.Printf("服务器运行在端口 %d\n", config.Port)
	fmt.Println("测试命令: ./iperf-go -c localhost -p 5209")

	// 5. 运行10秒后停止（实际使用中可以一直运行）
	time.Sleep(10 * time.Second)

	server.Stop()
	fmt.Printf("\n服务器已停止，处理了 %d 个测试\n", server.GetTestCount())
}

// useOriginalServerAPI 使用原有的Server API
func useOriginalServerAPI() {
	fmt.Println("📦 使用原有 Server API 的 StartContinuous 方法\n")

	// 1. 创建配置
	config := iperf.DefaultConfig()
	config.Role = iperf.RoleServer
	config.Port = 5210

	// 2. 创建服务器
	server, err := iperf.NewServer(config)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 3. 使用持续运行模式启动
	if err := server.StartContinuous(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}

	fmt.Printf("服务器运行在端口 %d\n", config.Port)

	// 运行一段时间
	time.Sleep(10 * time.Second)

	server.Stop()
	fmt.Println("服务器已停止")
}
