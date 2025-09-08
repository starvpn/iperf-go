package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"iperf-go/pkg/iperf"
)

func main() {
	// 解析命令行参数
	config := parseFlags()
	if config == nil {
		os.Exit(1)
	}

	// 根据角色运行
	if config.Role == iperf.RoleServer {
		runServer(config)
	} else {
		runClient(config)
	}
}

func parseFlags() *iperf.Config {
	// 命令行标志定义
	var helpFlag = flag.Bool("h", false, "this help")
	var serverFlag = flag.Bool("s", false, "server side")
	var clientFlag = flag.String("c", "", "client side (server address)")
	var reverseFlag = flag.Bool("R", false, "reverse mode. client receive, server send")
	var portFlag = flag.Uint("p", 5201, "connect/listen port")
	var protocolFlag = flag.String("proto", "tcp", "protocol under test")
	var durFlag = flag.Uint("d", 10, "duration (s)")
	var intervalFlag = flag.Uint("i", 1000, "test interval (ms)")
	var parallelFlag = flag.Uint("P", 1, "The number of simultaneous connections")
	var blksizeFlag = flag.Uint("l", 4*1024, "send/read block size")
	var bandwidthFlag = flag.String("b", "0", "bandwidth limit. (M/K), default MB/s")
	var debugFlag = flag.Bool("debug", false, "debug mode")
	var infoFlag = flag.Bool("info", false, "info mode")
	var noDelayFlag = flag.Bool("D", false, "no delay option")

	// RUDP 特定选项
	var sndWndFlag = flag.Uint("sw", 10, "rudp send window size")
	var rcvWndFlag = flag.Uint("rw", 512, "rudp receive window size")
	var readBufferSizeFlag = flag.Uint("rb", 4*1024, "read buffer size (Kb)")
	var writeBufferSizeFlag = flag.Uint("wb", 4*1024, "write buffer size (Kb)")
	var flushIntervalFlag = flag.Uint("f", 10, "flush interval for rudp (ms)")
	var noCongFlag = flag.Bool("nc", true, "no congestion control or BBR")
	var fastResendFlag = flag.Uint("fr", 0, "rudp fast resend strategy. 0 indicate turn off fast resend")
	var datashardsFlag = flag.Uint("data", 0, "rudp/kcp FEC dataShards option")
	var parityshardsFlag = flag.Uint("parity", 0, "rudp/kcp FEC parityShards option")

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return nil
	}

	// 创建配置
	config := iperf.DefaultConfig()

	// 设置角色
	if *serverFlag {
		config.Role = iperf.RoleServer
	} else if *clientFlag != "" {
		config.Role = iperf.RoleClient
		config.ServerAddr = *clientFlag
	} else {
		fmt.Println("Error: Must specify either -s (server) or -c <address> (client)")
		flag.Usage()
		return nil
	}

	// 基本配置
	config.Port = *portFlag
	config.Protocol = *protocolFlag
	config.Duration = time.Duration(*durFlag) * time.Second
	config.Interval = time.Duration(*intervalFlag) * time.Millisecond
	config.Reverse = *reverseFlag
	config.NoDelay = *noDelayFlag
	config.Parallel = *parallelFlag
	config.Blksize = *blksizeFlag

	// 解析带宽限制
	if *bandwidthFlag != "0" {
		bwStr := *bandwidthFlag
		multiplier := 1024 * 1024 // 默认 MB

		if len(bwStr) > 0 {
			suffix := bwStr[len(bwStr)-1]
			if suffix == 'M' || suffix == 'm' {
				multiplier = 1024 * 1024
				bwStr = bwStr[:len(bwStr)-1]
			} else if suffix == 'K' || suffix == 'k' {
				multiplier = 1024
				bwStr = bwStr[:len(bwStr)-1]
			}
		}

		if n, err := strconv.Atoi(bwStr); err == nil {
			config.Rate = uint(n * multiplier * 8) // 转换为 bits per second
			config.Burst = false
		} else {
			fmt.Printf("Error parsing bandwidth: %v\n", err)
			return nil
		}
	}

	// RUDP/KCP 配置
	config.SndWnd = *sndWndFlag
	config.RcvWnd = *rcvWndFlag
	config.ReadBufSize = *readBufferSizeFlag * 1024 // KB to bytes
	config.WriteBufSize = *writeBufferSizeFlag * 1024
	config.FlushInterval = *flushIntervalFlag
	config.NoCong = *noCongFlag
	config.FastResend = *fastResendFlag
	config.DataShards = *datashardsFlag
	config.ParityShards = *parityshardsFlag

	// 日志级别
	if *debugFlag {
		config.LogLevel = iperf.LogLevelDebug
	} else if *infoFlag {
		config.LogLevel = iperf.LogLevelInfo
	} else {
		config.LogLevel = iperf.LogLevelError
	}

	return config
}

func runServer(config *iperf.Config) {
	server, err := iperf.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 设置事件处理器
	server.SetEventHandler(func(event iperf.Event) {
		switch event.Type {
		case iperf.EventConnected:
			fmt.Println("Client connected")
		case iperf.EventInterval:
			// 间隔报告已由原始代码处理
		case iperf.EventComplete:
			fmt.Println("Test completed")
		case iperf.EventError:
			fmt.Printf("Error: %v\n", event.Error)
		}
	})

	fmt.Printf("Server listening on %d\n", config.Port)

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 等待中断信号
	select {}
}

func runClient(config *iperf.Config) {
	client, err := iperf.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 设置事件处理器
	client.SetEventHandler(func(event iperf.Event) {
		switch event.Type {
		case iperf.EventConnected:
			fmt.Printf("Connected to server %s:%d\n", config.ServerAddr, config.Port)
		case iperf.EventInterval:
			// 间隔报告已由原始代码处理
		case iperf.EventComplete:
			if result, ok := event.Data.(*iperf.TestResult); ok {
				printResult(result)
			}
		case iperf.EventError:
			fmt.Printf("Error: %v\n", event.Error)
		}
	})

	// 运行测试
	result, err := client.Run()
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	// 打印最终结果
	if result != nil {
		fmt.Println("\n--- Final Results ---")
		printResult(result)
	}
}

func printResult(result *iperf.TestResult) {
	fmt.Printf("Total Bytes: %.2f MB\n", float64(result.TotalBytes)/1024/1024)
	fmt.Printf("Duration: %.2f seconds\n", result.Duration.Seconds())
	fmt.Printf("Bandwidth: %.2f Mbps\n", result.Bandwidth)
	if result.RTT > 0 {
		fmt.Printf("RTT: %v\n", result.RTT)
	}
	if result.Retransmits > 0 {
		fmt.Printf("Retransmits: %d\n", result.Retransmits)
	}
	if result.PacketLoss > 0 {
		fmt.Printf("Packet Loss: %.2f%%\n", result.PacketLoss)
	}
}
