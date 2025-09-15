package iperf

import (
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	config := DefaultConfig()

	// 设置为服务器角色
	config.Role = RoleServer

	// 基本配置（使用 parseFlags 中的默认值）
	config.Port = 5201                        // 默认端口
	config.Protocol = "tcp"                   // 默认协议
	config.Duration = 10 * time.Second        // 默认持续时间 10秒
	config.Interval = 1000 * time.Millisecond // 默认间隔 1000ms
	config.Parallel = 1                       // 默认并行连接数
	config.Blksize = 4 * 1024                 // 默认块大小 4KB
	config.Reverse = false                    // 默认不启用反向模式
	config.NoDelay = false                    // 默认不启用 no delay
	config.Rate = 0                           // 默认无带宽限制
	config.Burst = true                       // 默认启用突发模式

	// RUDP/KCP 特定配置（使用 parseFlags 中的默认值）
	config.SndWnd = 10                    // 发送窗口大小
	config.RcvWnd = 512                   // 接收窗口大小
	config.ReadBufSize = 4 * 1024 * 1024  // 读缓冲区 4MB (4*1024 KB)
	config.WriteBufSize = 4 * 1024 * 1024 // 写缓冲区 4MB (4*1024 KB)
	config.FlushInterval = 10             // 刷新间隔 10ms
	config.NoCong = true                  // 默认禁用拥塞控制
	config.FastResend = 0                 // 默认关闭快速重传
	config.DataShards = 0                 // 默认无 FEC 数据分片
	config.ParityShards = 0               // 默认无 FEC 校验分片

	// 日志配置
	config.LogLevel = LogLevelError

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	err = server.RunTest()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	t.Log("Server test completed successfully")
}
