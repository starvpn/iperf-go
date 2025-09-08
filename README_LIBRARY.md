# iperf-go 库使用指南

## 概述

iperf-go 已被重构为一个可嵌入的 Go 库，可以轻松集成到其他应用程序中进行网络性能测试。

## 安装

```bash
go get github.com/yourusername/iperf-go
```

## 快速开始

### 作为客户端使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/yourusername/iperf-go/pkg/iperf"
)

func main() {
    // 创建客户端配置
    config := iperf.ClientConfig("192.168.1.100", 5201)
    config.Duration = 10 * time.Second
    config.Protocol = "tcp"
    
    // 创建并运行客户端
    client, err := iperf.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 运行测试
    result, err := client.Run()
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用结果
    fmt.Printf("带宽: %.2f Mbps\n", result.Bandwidth)
}
```

### 作为服务器使用

```go
package main

import (
    "fmt"
    "log"
    "github.com/yourusername/iperf-go/pkg/iperf"
)

func main() {
    // 创建服务器配置
    config := iperf.ServerConfig(5201)
    
    // 创建服务器
    server, err := iperf.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 启动服务器
    fmt.Println("服务器正在监听...")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
    
    // 保持运行
    select {}
}
```

## 主要 API

### 配置结构体

```go
type Config struct {
    // 基本配置
    Role       Role          // 角色：客户端或服务器
    ServerAddr string        // 服务器地址
    Port       uint          // 端口号
    Protocol   string        // 协议: tcp, udp, rudp, kcp
    Duration   time.Duration // 测试持续时间
    Interval   time.Duration // 报告间隔
    
    // 性能配置
    Parallel uint   // 并行连接数
    Blksize  uint   // 块大小
    Rate     uint   // 带宽限制
    
    // 高级配置
    Reverse  bool   // 反向模式
    NoDelay  bool   // TCP no delay
    LogLevel LogLevel // 日志级别
}
```

### 客户端 API

```go
// 创建客户端
func NewClient(config *Config) (*Client, error)

// 设置事件处理器
func (c *Client) SetEventHandler(handler EventHandler)

// 同步运行测试
func (c *Client) Run() (*TestResult, error)

// 异步运行测试
func (c *Client) RunAsync() error

// 停止测试
func (c *Client) Stop()

// 获取当前结果
func (c *Client) GetResult() *TestResult
```

### 服务器 API

```go
// 创建服务器
func NewServer(config *Config) (*Server, error)

// 设置事件处理器
func (s *Server) SetEventHandler(handler EventHandler)

// 启动服务器
func (s *Server) Start() error

// 停止服务器
func (s *Server) Stop()

// 检查运行状态
func (s *Server) IsRunning() bool
```

### 事件处理

```go
// 事件类型
const (
    EventConnected // 连接建立
    EventInterval  // 间隔报告
    EventComplete  // 测试完成
    EventError     // 发生错误
)

// 设置事件处理器
client.SetEventHandler(func(event iperf.Event) {
    switch event.Type {
    case iperf.EventInterval:
        // 处理间隔报告
    case iperf.EventComplete:
        // 处理测试完成
    }
})
```

### 测试结果

```go
type TestResult struct {
    TotalBytes      uint64        // 总传输字节数
    Duration        time.Duration // 实际测试时长
    Bandwidth       float64       // 平均带宽 (Mbps)
    RTT             time.Duration // 平均往返时间
    PacketLoss      float64       // 丢包率 (%)
    Retransmits     uint          // 重传次数
    IntervalResults []IntervalResult // 间隔结果
}
```

## 高级用法

### 1. 使用不同协议

```go
// TCP (默认)
config.Protocol = "tcp"

// UDP
config.Protocol = "udp"

// RUDP (可靠 UDP)
config.Protocol = "rudp"
config.SndWnd = 512
config.RcvWnd = 512

// KCP
config.Protocol = "kcp"
config.DataShards = 10
config.ParityShards = 3
```

### 2. 自定义日志

```go
// 实现 Logger 接口
type MyLogger struct {
    // ...
}

func (l *MyLogger) Error(v ...interface{}) {
    // 自定义错误日志
}

// 使用自定义日志
config.Logger = &MyLogger{}
```

### 3. 在 Web 服务中集成

```go
func HandleSpeedTest(w http.ResponseWriter, r *http.Request) {
    serverAddr := r.URL.Query().Get("server")
    
    config := iperf.ClientConfig(serverAddr, 5201)
    config.Duration = 5 * time.Second
    
    client, _ := iperf.NewClient(config)
    result, _ := client.Run()
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "bandwidth": result.Bandwidth,
        "rtt": result.RTT,
    })
}
```

### 4. 批量测试

```go
func TestMultipleServers(servers []string) {
    results := make(map[string]float64)
    
    for _, server := range servers {
        config := iperf.ClientConfig(server, 5201)
        config.Duration = 10 * time.Second
        
        client, _ := iperf.NewClient(config)
        result, _ := client.Run()
        
        results[server] = result.Bandwidth
    }
    
    // 分析结果...
}
```

## 迁移指南

### 从命令行工具迁移

如果你之前使用命令行工具：

```bash
./iperf-go -c 192.168.1.100 -p 5201 -d 10 -P 5
```

现在可以使用库：

```go
config := iperf.ClientConfig("192.168.1.100", 5201)
config.Duration = 10 * time.Second
config.Parallel = 5

client, _ := iperf.NewClient(config)
result, _ := client.Run()
```

### 从原始 API 迁移

如果你之前直接使用 IperfTest：

```go
// 旧代码
test := iperf.NewIperfTest()
test.Init()
test.ParseArguments()
test.RunTest()
```

现在使用新 API：

```go
// 新代码
config := iperf.DefaultConfig()
client, _ := iperf.NewClient(config)
result, _ := client.Run()
```

## 注意事项

1. **并发安全**：Client 和 Server 实例都是线程安全的
2. **资源清理**：记得调用 `Stop()` 方法清理资源
3. **错误处理**：始终检查返回的错误
4. **超时设置**：可以使用 context 控制超时

## 示例项目

查看 `examples/` 目录获取更多使用示例：

- `basic_usage.go` - 基本使用示例
- `web_integration.go` - Web 服务集成示例
- `monitoring.go` - 监控系统集成示例
- `benchmark.go` - 性能基准测试示例

## 许可证

MIT License
