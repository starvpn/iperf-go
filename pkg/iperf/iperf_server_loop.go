package iperf

import (
	"fmt"
	"time"
)

// RunServerLoop 运行持续服务的服务器（简单实现）
// 这是一个简单的包装函数，通过循环调用原有的runServer来实现持续服务
func (test *IperfTest) RunServerLoop() int {
	if !test.isServer {
		Log.Error("RunServerLoop只能在服务器模式下使用")
		return -1
	}

	fmt.Printf("服务器持续运行模式启动，监听端口 %d\n", test.port)
	fmt.Println("服务器将持续接受客户端连接...")
	fmt.Println("按 Ctrl+C 退出\n")

	testCount := 0
	consecutiveErrors := 0
	maxConsecutiveErrors := 5

	for {
		testCount++
		fmt.Printf("\n[测试 #%d] 等待客户端连接...\n", testCount)

		// 运行一次服务器测试
		rtn := test.runServer()

		if rtn < 0 {
			consecutiveErrors++
			fmt.Printf("[测试 #%d] 测试失败，错误码: %d\n", testCount, rtn)

			// 如果连续失败太多次，可能有严重问题
			if consecutiveErrors >= maxConsecutiveErrors {
				Log.Errorf("连续失败 %d 次，停止服务器", consecutiveErrors)
				return rtn
			}

			// 根据错误类型决定等待时间
			switch rtn {
			case -1: // 监听失败
				fmt.Println("监听失败，可能端口被占用，等待5秒...")
				time.Sleep(5 * time.Second)
			case -2: // Accept失败
				fmt.Println("接受连接失败，等待1秒...")
				time.Sleep(1 * time.Second)
			default:
				fmt.Println("其他错误，等待2秒...")
				time.Sleep(2 * time.Second)
			}
		} else {
			// 成功完成一次测试
			consecutiveErrors = 0 // 重置连续错误计数
			fmt.Printf("[测试 #%d] 测试成功完成\n", testCount)

			// 显示一些统计
			if test.bytesReceived > 0 || test.bytesSent > 0 {
				fmt.Printf("  接收: %.2f MB, 发送: %.2f MB\n",
					float64(test.bytesReceived)/1024/1024,
					float64(test.bytesSent)/1024/1024)
			}

			// 重置测试状态以准备下一次
			test.resetForNextTest()

			// 短暂等待以确保资源释放
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// resetForNextTest 重置测试状态以准备下一次测试
func (test *IperfTest) resetForNextTest() {
	// 清理流
	for _, sp := range test.streams {
		if sp != nil && sp.conn != nil {
			sp.conn.Close()
		}
	}
	test.streams = nil

	// 重置统计
	test.bytesReceived = 0
	test.bytesSent = 0
	test.blocksReceived = 0
	test.blocksSent = 0

	// 重置状态
	test.state = IPERF_START
	test.done = false

	// 关闭并重置连接
	if test.ctrlConn != nil {
		test.ctrlConn.Close()
		test.ctrlConn = nil
	}

	// 关闭协议监听器（如果有）
	if test.protoListener != nil {
		test.protoListener.Close()
		test.protoListener = nil
	}

	// 关闭主监听器（如果有）
	if test.listener != nil {
		test.listener.Close()
		test.listener = nil
	}

	// 重置定时器
	if test.timer.timer != nil {
		test.timer.timer.Stop()
	}
	test.timer = ITimer{}

	if test.statsTicker.ticker != nil {
		test.statsTicker.ticker.Stop()
	}
	test.statsTicker = ITicker{}

	if test.reportTicker.ticker != nil {
		test.reportTicker.ticker.Stop()
	}
	test.reportTicker = ITicker{}

	// 重新初始化协议（保持原有协议设置）
	if test.proto != nil {
		protoName := test.proto.name()
		for _, p := range test.protocols {
			if p.name() == protoName {
				test.proto = p
				break
			}
		}
	}
}
