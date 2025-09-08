package iperf

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// TestResult 包含测试结果
type TestResult struct {
	TotalBytes      uint64           // 总传输字节数
	Duration        time.Duration    // 实际测试时长
	Bandwidth       float64          // 平均带宽 (Mbps)
	RTT             time.Duration    // 平均往返时间
	PacketLoss      float64          // 丢包率 (%)
	Retransmits     uint             // 重传次数
	IntervalResults []IntervalResult // 间隔结果
}

// IntervalResult 包含每个间隔的结果
type IntervalResult struct {
	StartTime   time.Time
	EndTime     time.Time
	Bytes       uint64
	Bandwidth   float64 // Mbps
	RTT         time.Duration
	Retransmits uint
}

// EventType 定义事件类型
type EventType int

const (
	EventConnected EventType = iota
	EventInterval
	EventComplete
	EventError
)

// Event 定义事件
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      interface{}
	Error     error
}

// EventHandler 事件处理函数
type EventHandler func(event Event)

// Client 是 iperf 客户端的新 API
type Client struct {
	config       *Config
	test         *IperfTest
	eventHandler EventHandler
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	result       *TestResult
}

// NewClient 创建新的客户端
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化内部测试对象
	if err := client.initTest(); err != nil {
		return nil, err
	}

	return client, nil
}

// SetEventHandler 设置事件处理器
func (c *Client) SetEventHandler(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandler = handler
}

// initTest 初始化内部测试对象
func (c *Client) initTest() error {
	c.test = NewIperfTest()
	if c.test == nil {
		return errors.New("failed to create test instance")
	}

	// 先初始化协议列表
	c.test.Init()

	// 然后应用配置（包括设置协议）
	c.applyConfig()

	return nil
}

// applyConfig 应用配置到测试对象
func (c *Client) applyConfig() {
	c.test.isServer = (c.config.Role == RoleServer)
	c.test.addr = c.config.ServerAddr
	c.test.port = c.config.Port
	c.test.duration = uint(c.config.Duration.Seconds())
	c.test.interval = uint(c.config.Interval.Milliseconds())
	c.test.reverse = c.config.Reverse
	c.test.noDelay = c.config.NoDelay
	c.test.streamNum = c.config.Parallel

	// 设置协议
	c.test.setProtocol(c.config.Protocol)

	// 应用设置
	c.test.setting.blksize = c.config.Blksize
	c.test.setting.burst = c.config.Burst
	c.test.setting.rate = c.config.Rate
	c.test.setting.sndWnd = c.config.SndWnd
	c.test.setting.rcvWnd = c.config.RcvWnd
	c.test.setting.readBufSize = c.config.ReadBufSize
	c.test.setting.writeBufSize = c.config.WriteBufSize
	c.test.setting.flushInterval = c.config.FlushInterval
	c.test.setting.noCong = c.config.NoCong
	c.test.setting.fastResend = c.config.FastResend
	c.test.setting.dataShards = c.config.DataShards
	c.test.setting.parityShards = c.config.ParityShards

	// 设置模式
	c.test.setTestReverse(c.config.Reverse)
}

// Run 运行测试（阻塞直到完成）
func (c *Client) Run() (*TestResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 发送连接事件
	c.emitEvent(Event{
		Type:      EventConnected,
		Timestamp: time.Now(),
	})

	// 设置结果回调
	c.test.statsCallback = c.statsCallback
	c.test.reporterCallback = c.reporterCallback

	// 运行测试
	if rtn := c.test.RunTest(); rtn < 0 {
		err := fmt.Errorf("test failed with code: %d", rtn)
		c.emitEvent(Event{
			Type:      EventError,
			Timestamp: time.Now(),
			Error:     err,
		})
		return nil, err
	}

	// 收集结果
	c.collectResults()

	// 发送完成事件
	c.emitEvent(Event{
		Type:      EventComplete,
		Timestamp: time.Now(),
		Data:      c.result,
	})

	return c.result, nil
}

// RunAsync 异步运行测试
func (c *Client) RunAsync() error {
	go func() {
		_, _ = c.Run()
	}()
	return nil
}

// Stop 停止测试
func (c *Client) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	c.test.done = true
	c.test.FreeTest()
}

// GetResult 获取当前结果
func (c *Client) GetResult() *TestResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.result
}

// statsCallback 统计回调
func (c *Client) statsCallback(test *IperfTest) {
	// 收集统计信息
	iperfStatsCallback(test)
}

// reporterCallback 报告回调
func (c *Client) reporterCallback(test *IperfTest) {
	// 发送间隔事件
	c.emitEvent(Event{
		Type:      EventInterval,
		Timestamp: time.Now(),
		Data:      c.collectIntervalResult(),
	})

	// 调用原始报告回调
	iperfReporterCallback(test)
}

// collectResults 收集最终结果
func (c *Client) collectResults() {
	c.result = &TestResult{
		TotalBytes:      c.test.bytesSent + c.test.bytesReceived,
		Duration:        time.Duration(c.test.duration) * time.Second,
		IntervalResults: c.collectAllIntervalResults(),
	}

	// 计算平均带宽
	if c.result.Duration > 0 {
		c.result.Bandwidth = float64(c.result.TotalBytes*8) / c.result.Duration.Seconds() / 1000000
	}
}

// collectIntervalResult 收集间隔结果
func (c *Client) collectIntervalResult() *IntervalResult {
	// TODO: 实现间隔结果收集
	return &IntervalResult{}
}

// collectAllIntervalResults 收集所有间隔结果
func (c *Client) collectAllIntervalResults() []IntervalResult {
	// TODO: 实现所有间隔结果收集
	return []IntervalResult{}
}

// emitEvent 发送事件
func (c *Client) emitEvent(event Event) {
	if c.eventHandler != nil {
		c.eventHandler(event)
	}
}

// Server 是 iperf 服务器的新 API
type Server struct {
	config       *Config
	test         *IperfTest
	eventHandler EventHandler
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
}

// NewServer 创建新的服务器
func NewServer(config *Config) (*Server, error) {
	if config == nil {
		config = ServerConfig(5201)
	}

	config.Role = RoleServer

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化内部测试对象
	if err := server.initTest(); err != nil {
		return nil, err
	}

	return server, nil
}

// SetEventHandler 设置事件处理器
func (s *Server) SetEventHandler(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventHandler = handler
}

// initTest 初始化内部测试对象
func (s *Server) initTest() error {
	s.test = NewIperfTest()
	if s.test == nil {
		return errors.New("failed to create test instance")
	}

	// 先初始化协议列表
	s.test.Init()

	// 然后应用配置
	s.applyConfig()

	return nil
}

// applyConfig 应用配置到测试对象
func (s *Server) applyConfig() {
	s.test.isServer = true
	s.test.port = s.config.Port
	s.test.duration = uint(s.config.Duration.Seconds())
	s.test.interval = uint(s.config.Interval.Milliseconds())
	s.test.reverse = s.config.Reverse
	s.test.noDelay = s.config.NoDelay

	// 应用设置
	s.test.setting.blksize = s.config.Blksize
	s.test.setting.burst = s.config.Burst
	s.test.setting.rate = s.config.Rate
	s.test.setting.sndWnd = s.config.SndWnd
	s.test.setting.rcvWnd = s.config.RcvWnd
	s.test.setting.readBufSize = s.config.ReadBufSize
	s.test.setting.writeBufSize = s.config.WriteBufSize
	s.test.setting.flushInterval = s.config.FlushInterval
	s.test.setting.noCong = s.config.NoCong
	s.test.setting.fastResend = s.config.FastResend
	s.test.setting.dataShards = s.config.DataShards
	s.test.setting.parityShards = s.config.ParityShards

	// 设置模式
	s.test.setTestReverse(s.config.Reverse)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server already running")
	}

	s.running = true

	// 设置回调
	s.test.statsCallback = iperfStatsCallback
	s.test.reporterCallback = iperfReporterCallback

	// 在新协程中运行服务器
	go func() {
		if rtn := s.test.RunTest(); rtn < 0 {
			s.emitEvent(Event{
				Type:      EventError,
				Timestamp: time.Now(),
				Error:     fmt.Errorf("server failed with code: %d", rtn),
			})
		}
		s.running = false
	}()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.test.done = true
	s.test.FreeTest()
	s.running = false
}

// IsRunning 检查服务器是否正在运行
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// emitEvent 发送事件
func (s *Server) emitEvent(event Event) {
	if s.eventHandler != nil {
		s.eventHandler(event)
	}
}
