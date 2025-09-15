package iperf

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ContinuousServer 提供简单的持续运行服务器API
type ContinuousServer struct {
	config       *Config
	running      bool
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	eventHandler EventHandler
	testCount    int
	currentTest  *IperfTest // 保存当前运行的测试实例
}

// NewContinuousServer 创建持续运行服务器
func NewContinuousServer(config *Config) (*ContinuousServer, error) {
	if config == nil {
		config = ServerConfig(5201)
	}

	config.Role = RoleServer

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ContinuousServer{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// SetEventHandler 设置事件处理器
func (s *ContinuousServer) SetEventHandler(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventHandler = handler
}

// Start 启动持续运行服务器
func (s *ContinuousServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("server already running")
	}
	s.running = true
	s.mu.Unlock()

	// 在新协程中运行服务器循环
	go s.runLoop()

	return nil
}

// runLoop 服务器主循环
func (s *ContinuousServer) runLoop() {
	fmt.Printf("持续运行服务器启动，监听端口 %d\n", s.config.Port)

	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("服务器停止...")
			return
		default:
		}

		s.mu.Lock()
		s.testCount++
		testNum := s.testCount
		s.mu.Unlock()

		// 发送事件
		s.emitEvent(Event{
			Type:      EventConnected,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"test_num": testNum},
		})

		fmt.Printf("\n[测试 #%d] 等待客户端连接...\n", testNum)

		// 创建新的测试实例
		test := NewIperfTest()
		if test == nil {
			s.emitEvent(Event{
				Type:      EventError,
				Timestamp: time.Now(),
				Error:     errors.New("failed to create test instance"),
			})
			time.Sleep(1 * time.Second)
			continue
		}

		// 初始化
		test.Init()

		// 应用配置
		s.applyConfig(test)

		// 保存当前测试实例
		s.mu.Lock()
		s.currentTest = test
		s.mu.Unlock()

		// 在新协程中运行测试，以便能响应停止信号
		testDone := make(chan int)
		go func() {
			rtn := test.runServer()
			testDone <- rtn
		}()

		// 等待测试完成或停止信号
		var rtn int
		startTime := time.Now()

		select {
		case <-s.ctx.Done():
			// 收到停止信号，关闭当前测试
			fmt.Printf("[测试 #%d] 收到停止信号，关闭测试...\n", testNum)
			s.cleanupTest(test)
			return

		case rtn = <-testDone:
			// 测试正常完成
			duration := time.Since(startTime)

			if rtn < 0 {
				fmt.Printf("[测试 #%d] 测试失败，错误码: %d\n", testNum, rtn)
				s.emitEvent(Event{
					Type:      EventError,
					Timestamp: time.Now(),
					Error:     fmt.Errorf("test failed with code: %d", rtn),
					Data:      map[string]interface{}{"test_num": testNum},
				})

				// 根据错误类型决定等待时间
				if rtn == -1 {
					time.Sleep(2 * time.Second) // 监听失败
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			} else {
				fmt.Printf("[测试 #%d] 测试成功完成 (耗时: %.2f秒)\n", testNum, duration.Seconds())

				// 创建测试结果
				result := &TestResult{
					TotalBytes: test.bytesReceived + test.bytesSent,
					Duration:   duration,
					Bandwidth:  float64((test.bytesReceived+test.bytesSent)*8) / duration.Seconds() / 1000000, // Mbps
				}

				s.emitEvent(Event{
					Type:      EventComplete,
					Timestamp: time.Now(),
					Data:      result,
				})
			}

			// 清理资源
			s.cleanupTest(test)

			// 清除当前测试引用
			s.mu.Lock()
			s.currentTest = nil
			s.mu.Unlock()

			// 短暂等待
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// applyConfig 应用配置到测试实例
func (s *ContinuousServer) applyConfig(test *IperfTest) {
	test.isServer = true
	test.port = s.config.Port
	test.duration = uint(s.config.Duration.Seconds())
	test.interval = uint(s.config.Interval.Milliseconds())
	test.reverse = s.config.Reverse
	test.noDelay = s.config.NoDelay

	// 应用设置
	if test.setting != nil {
		test.setting.blksize = s.config.Blksize
		test.setting.burst = s.config.Burst
		test.setting.rate = s.config.Rate
		test.setting.sndWnd = s.config.SndWnd
		test.setting.rcvWnd = s.config.RcvWnd
		test.setting.readBufSize = s.config.ReadBufSize
		test.setting.writeBufSize = s.config.WriteBufSize
		test.setting.flushInterval = s.config.FlushInterval
		test.setting.noCong = s.config.NoCong
		test.setting.fastResend = s.config.FastResend
		test.setting.dataShards = s.config.DataShards
		test.setting.parityShards = s.config.ParityShards
	}

	// 设置模式
	test.setTestReverse(s.config.Reverse)

	// 设置回调
	test.statsCallback = iperfStatsCallback
	test.reporterCallback = iperfReporterCallback
}

// cleanupTest 清理测试资源
func (s *ContinuousServer) cleanupTest(test *IperfTest) {
	// 关闭所有流
	for _, sp := range test.streams {
		if sp != nil && sp.conn != nil {
			sp.conn.Close()
		}
	}

	// 关闭连接
	if test.ctrlConn != nil {
		test.ctrlConn.Close()
	}

	// 关闭监听器
	if test.listener != nil {
		test.listener.Close()
	}

	if test.protoListener != nil {
		test.protoListener.Close()
	}

	// 停止定时器
	if test.timer.timer != nil {
		test.timer.timer.Stop()
	}

	if test.statsTicker.ticker != nil {
		test.statsTicker.ticker.Stop()
	}

	if test.reportTicker.ticker != nil {
		test.reportTicker.ticker.Stop()
	}
}

// Stop 停止服务器
func (s *ContinuousServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false

	// 如果有当前运行的测试，立即关闭它的监听器
	if s.currentTest != nil {
		fmt.Println("关闭当前运行的测试监听器...")
		if s.currentTest.listener != nil {
			s.currentTest.listener.Close()
		}
		if s.currentTest.protoListener != nil {
			s.currentTest.protoListener.Close()
		}
	}

	// 发送取消信号
	s.cancel()

	fmt.Printf("服务器已停止，共处理了 %d 次测试\n", s.testCount)
}

// IsRunning 检查服务器是否正在运行
func (s *ContinuousServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetTestCount 获取已处理的测试数量
func (s *ContinuousServer) GetTestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.testCount
}

// emitEvent 发送事件
func (s *ContinuousServer) emitEvent(event Event) {
	if s.eventHandler != nil {
		s.eventHandler(event)
	}
}
