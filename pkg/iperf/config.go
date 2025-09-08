package iperf

import (
	"time"
)

// Config 是 iperf 测试的配置结构体
type Config struct {
	// 基本配置
	Role       Role          // 角色：客户端或服务器
	ServerAddr string        // 服务器地址（客户端模式需要）
	Port       uint          // 端口号
	Protocol   string        // 协议类型: tcp, udp, rudp, kcp
	Duration   time.Duration // 测试持续时间
	Interval   time.Duration // 报告间隔
	Reverse    bool          // 反向模式
	NoDelay    bool          // TCP no delay 选项

	// 连接配置
	Parallel uint // 并行连接数
	Blksize  uint // 块大小
	Burst    bool // 突发模式
	Rate     uint // 带宽限制 (bits per second)

	// RUDP/KCP 特定配置
	SndWnd        uint // 发送窗口大小
	RcvWnd        uint // 接收窗口大小
	ReadBufSize   uint // 读缓冲区大小 (bytes)
	WriteBufSize  uint // 写缓冲区大小 (bytes)
	FlushInterval uint // 刷新间隔 (ms)
	NoCong        bool // 禁用拥塞控制
	FastResend    uint // 快速重传策略
	DataShards    uint // FEC 数据分片
	ParityShards  uint // FEC 校验分片

	// 日志配置
	LogLevel LogLevel // 日志级别
	Logger   Logger   // 自定义日志记录器（可选）
}

// Role 定义测试角色
type Role int

const (
	RoleServer Role = iota
	RoleClient
)

// LogLevel 定义日志级别
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// Logger 定义日志接口
type Logger interface {
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Role:          RoleClient,
		ServerAddr:    "127.0.0.1",
		Port:          5201,
		Protocol:      "tcp",
		Duration:      10 * time.Second,
		Interval:      1 * time.Second,
		Parallel:      1,
		Blksize:       DEFAULT_TCP_BLKSIZE,
		Burst:         true,
		SndWnd:        10,
		RcvWnd:        512,
		ReadBufSize:   4 * 1024 * 1024,
		WriteBufSize:  4 * 1024 * 1024,
		FlushInterval: 10,
		NoCong:        true,
		LogLevel:      LogLevelError,
	}
}

// ServerConfig 创建服务器默认配置
func ServerConfig(port uint) *Config {
	config := DefaultConfig()
	config.Role = RoleServer
	config.Port = port
	return config
}

// ClientConfig 创建客户端默认配置
func ClientConfig(serverAddr string, port uint) *Config {
	config := DefaultConfig()
	config.Role = RoleClient
	config.ServerAddr = serverAddr
	config.Port = port
	return config
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// TODO: 添加配置验证逻辑
	return nil
}
