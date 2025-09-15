# iperf-go API库使用指南 - 持续运行服务器

## 概述

iperf-go 提供了完整的API库支持，可以轻松在您的Go应用中集成网络性能测试功能。现在API库也支持持续运行模式，服务器可以连续处理多个客户端测试。

## 快速开始

### 1. 最简单的例子

```go
package main

import (
    "log"
    "time"
    "iperf-go/pkg/iperf"
)

func main() {
    // 创建配置
    config := iperf.ServerConfig(5201)
    
    // 创建持续运行服务器
    server, err := iperf.NewContinuousServer(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 启动服务器
    server.Start()
    
    // 服务器会持续运行，处理所有客户端连接
    select {} // 保持运行
}
```

## API方式对比

### 方式1: ContinuousServer（推荐）

```go
// 新的持续运行API
server, _ := iperf.NewContinuousServer(config)
server.Start() // 自动循环，持续接受连接
```

**优点**：
- ✅ 专为持续运行设计
- ✅ 简单易用
- ✅ 自动资源管理
- ✅ 内置事件系统

### 方式2: 原有Server API + StartContinuous

```go
// 使用原有API的持续运行模式
server, _ := iperf.NewServer(config)
server.StartContinuous() // 持续运行模式
```

**优点**：
- ✅ 向后兼容
- ✅ 使用熟悉的API

### 方式3: 手动循环（不推荐）

```go
// 手动循环调用
for {
    test := iperf.NewIperfTest()
    test.Init()
    test.RunTest()
    test.FreeTest()
}
```

## 完整示例

### 带事件处理的服务器

```go
package main

import (
    "fmt"
    "log"
    "time"
    "iperf-go/pkg/iperf"
)

func main() {
    // 1. 创建配置
    config := &iperf.Config{
        Role:     iperf.RoleServer,
        Port:     5201,
        Protocol: "tcp",
        Duration: 10 * time.Second,
        Interval: 1 * time.Second,
    }
    
    // 2. 创建服务器
    server, err := iperf.NewContinuousServer(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 设置事件处理器
    var testCount int
    server.SetEventHandler(func(event iperf.Event) {
        switch event.Type {
        case iperf.EventConnected:
            testCount++
            fmt.Printf("客户端 #%d 已连接\n", testCount)
            
        case iperf.EventComplete:
            if result, ok := event.Data.(*iperf.TestResult); ok {
                fmt.Printf("测试完成: %.2f Mbps\n", result.Bandwidth)
            }
            
        case iperf.EventError:
            fmt.Printf("错误: %v\n", event.Error)
        }
    })
    
    // 4. 启动服务器
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("服务器运行在端口 %d\n", config.Port)
    
    // 5. 保持运行
    select {}
}
```

## 配置选项

```go
config := &iperf.Config{
    // 基本配置
    Role:       iperf.RoleServer,     // 角色
    Port:       5201,                  // 端口
    Protocol:   "tcp",                 // 协议: tcp/udp/rudp/kcp
    
    // 性能配置
    Duration:   10 * time.Second,     // 测试时长
    Interval:   1 * time.Second,       // 报告间隔
    Parallel:   1,                     // 并行连接数
    Blksize:    128 * 1024,          // 块大小
    
    // 高级配置
    Reverse:    false,                // 反向模式
    NoDelay:    false,                // TCP NoDelay
    Rate:       0,                    // 带宽限制(0=无限制)
    
    // RUDP/KCP配置
    SndWnd:     512,                  // 发送窗口
    RcvWnd:     512,                  // 接收窗口
}
```

## 事件系统

### 事件类型

```go
const (
    EventConnected  // 客户端连接
    EventInterval   // 间隔报告
    EventComplete   // 测试完成
    EventError      // 发生错误
)
```

### 事件数据

```go
type Event struct {
    Type      EventType     // 事件类型
    Timestamp time.Time     // 时间戳
    Data      interface{}   // 事件数据（取决于类型）
    Error     error         // 错误信息（仅EventError）
}

type TestResult struct {
    TotalBytes  uint64        // 总字节数
    Duration    time.Duration // 测试时长
    Bandwidth   float64       // 带宽(Mbps)
    RTT         time.Duration // 往返时间
    PacketLoss  float64       // 丢包率(%)
    Retransmits uint          // 重传次数
}
```

## 实际应用场景

### 1. Web服务集成

```go
// 在HTTP服务中提供速度测试API
func SpeedTestHandler(w http.ResponseWriter, r *http.Request) {
    config := iperf.ClientConfig("speedtest.server.com", 5201)
    config.Duration = 5 * time.Second
    
    client, _ := iperf.NewClient(config)
    result, _ := client.Run()
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "bandwidth": result.Bandwidth,
        "rtt": result.RTT,
    })
}
```

### 2. 监控系统集成

```go
// 定期测试网络性能
func NetworkMonitor() {
    server, _ := iperf.NewContinuousServer(config)
    
    server.SetEventHandler(func(event iperf.Event) {
        if event.Type == iperf.EventComplete {
            // 发送到监控系统
            metrics.RecordBandwidth(result.Bandwidth)
            metrics.RecordRTT(result.RTT)
        }
    })
    
    server.Start()
}
```

### 3. 自动化测试

```go
// 批量测试多个服务器
func TestMultipleServers(servers []string) {
    for _, server := range servers {
        config := iperf.ClientConfig(server, 5201)
        client, _ := iperf.NewClient(config)
        result, _ := client.Run()
        
        fmt.Printf("%s: %.2f Mbps\n", server, result.Bandwidth)
    }
}
```

## 编译和运行示例

```bash
# 编译API服务器示例
go build -o api_server examples/api_continuous_server.go

# 运行服务器
./api_server

# 在另一个终端测试
./iperf-go -c localhost -p 5208
```

## API方法参考

### ContinuousServer

```go
// 创建服务器
func NewContinuousServer(config *Config) (*ContinuousServer, error)

// 设置事件处理器
func (s *ContinuousServer) SetEventHandler(handler EventHandler)

// 启动服务器
func (s *ContinuousServer) Start() error

// 停止服务器
func (s *ContinuousServer) Stop()

// 检查运行状态
func (s *ContinuousServer) IsRunning() bool

// 获取处理的测试数
func (s *ContinuousServer) GetTestCount() int
```

### 原有Server API

```go
// 创建服务器
func NewServer(config *Config) (*Server, error)

// 单次测试模式
func (s *Server) Start() error

// 持续运行模式
func (s *Server) StartContinuous() error

// 停止服务器
func (s *Server) Stop()
```

## 注意事项

1. **并发安全**：ContinuousServer是线程安全的
2. **资源管理**：服务器会自动管理资源，无需手动清理
3. **错误处理**：通过事件系统处理错误
4. **优雅关闭**：调用Stop()方法优雅关闭

## 性能考虑

- 每个测试使用独立的资源
- 测试间有短暂间隔（200ms）确保资源释放
- 支持高并发（受系统资源限制）

## 故障排除

### 问题：端口被占用
```go
// 解决：使用不同端口或等待释放
config.Port = 5202 // 使用其他端口
```

### 问题：内存增长
```go
// 解决：确保调用Stop()释放资源
defer server.Stop()
```

### 问题：测试失败
```go
// 解决：通过事件处理器查看错误
server.SetEventHandler(func(event iperf.Event) {
    if event.Type == iperf.EventError {
        log.Printf("错误详情: %v", event.Error)
    }
})
```

## 总结

iperf-go的API库现已完全支持持续运行模式：

- ✅ **ContinuousServer API** - 专为持续运行设计
- ✅ **事件驱动架构** - 灵活的事件处理
- ✅ **自动资源管理** - 无需手动清理
- ✅ **线程安全** - 支持并发使用
- ✅ **向后兼容** - 原有API仍可使用

无论是简单的速度测试还是复杂的网络监控系统，iperf-go API都能满足您的需求。
