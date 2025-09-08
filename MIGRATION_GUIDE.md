# iperf-go 包改造迁移指南

## 项目结构变更

### 改造前
```
iperf-go/
├── cmd/
│   └── main.go          # 命令行入口
├── pkg/
│   └── iperf/          # 核心代码
└── README.md
```

### 改造后
```
iperf-go/
├── cmd/
│   ├── main.go         # 旧的命令行入口（保留兼容）
│   └── iperf-cli/      # 新的命令行工具
│       └── main.go
├── pkg/
│   └── iperf/
│       ├── config.go         # 新增：配置结构体
│       ├── iperf_new_api.go  # 新增：友好的 API
│       └── ...               # 原有文件
├── examples/                 # 新增：使用示例
│   └── basic_usage.go
├── README.md                 # 原有文档
├── README_LIBRARY.md         # 新增：库使用文档
└── MIGRATION_GUIDE.md        # 本文件
```

## 主要改进

### 1. 配置管理
- **改造前**：通过命令行参数配置
- **改造后**：使用 `Config` 结构体，支持编程配置

### 2. API 设计
- **改造前**：需要按顺序调用多个方法
- **改造后**：简化的 API，一次调用即可完成测试

### 3. 错误处理
- **改造前**：返回错误码，通过日志输出
- **改造后**：返回 Go 标准的 error 对象

### 4. 结果获取
- **改造前**：结果只能通过日志查看
- **改造后**：返回结构化的 `TestResult` 对象

### 5. 事件机制
- **改造前**：无事件通知
- **改造后**：支持事件回调，实时获取测试进度

## 使用步骤

### 步骤 1：安装依赖
```bash
# 更新 go.mod
go mod tidy
```

### 步骤 2：构建库
```bash
# 构建包
go build ./pkg/iperf
```

### 步骤 3：构建命令行工具（可选）
```bash
# 构建新的命令行工具
go build -o iperf-cli ./cmd/iperf-cli

# 或构建原有的命令行工具（保持兼容）
go build -o iperf-go ./cmd
```

### 步骤 4：运行测试
```bash
# 运行单元测试
go test ./pkg/iperf

# 运行示例
go run examples/basic_usage.go
```

## 集成到现有项目

### 1. 添加依赖
```go
import "github.com/yourusername/iperf-go/pkg/iperf"
```

### 2. 创建配置
```go
config := iperf.ClientConfig("server.example.com", 5201)
config.Duration = 10 * time.Second
```

### 3. 运行测试
```go
client, err := iperf.NewClient(config)
if err != nil {
    return err
}

result, err := client.Run()
if err != nil {
    return err
}

fmt.Printf("Bandwidth: %.2f Mbps\n", result.Bandwidth)
```

## 兼容性说明

### 向后兼容
- 原有的 `IperfTest` 结构体和方法仍然保留
- 原有的命令行工具仍可正常使用
- 不影响现有的协议实现

### 破坏性变更
- 无，所有改进都是新增功能

## 下一步改进建议

### 短期改进
1. 完善错误处理，返回更详细的错误信息
2. 添加更多的配置验证
3. 实现完整的间隔结果收集
4. 添加单元测试

### 长期改进
1. 支持更多协议（QUIC、WebRTC 等）
2. 添加 gRPC API 支持
3. 实现分布式测试功能
4. 提供 Web UI 界面
5. 支持配置文件（YAML/JSON）

## 注意事项

1. **日志处理**：新 API 支持自定义日志接口，建议实现自己的日志处理器
2. **并发测试**：可以同时创建多个客户端实例进行并发测试
3. **资源清理**：记得调用 `Stop()` 方法释放资源
4. **超时控制**：可以使用 context 实现超时控制

## 问题反馈

如有问题或建议，请提交 Issue 或 Pull Request。
