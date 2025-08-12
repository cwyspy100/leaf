# chanrpc.go 实现说明文档

## 概述

`chanrpc.go` 是一个基于 Go 通道（channel）的 RPC（远程过程调用）实现，提供了轻量级的协程间通信机制。该实现通过 channel 实现了函数注册、调用和结果返回的完整流程。

## 核心组件

### 1. Server（服务器端）

#### 结构体定义
```go
type Server struct {
    functions map[interface{}]interface{}  // 存储注册的函数
    ChanCall  chan *CallInfo               // 接收调用请求的通道
}
```

#### 功能特点
- **单协程模型**：每个 Server 实例必须在单独的 goroutine 中使用（非协程安全）
- **函数注册**：支持三种类型的函数签名
  - `func(args []interface{})` - 无返回值
  - `func(args []interface{}) interface{}` - 返回单个值
  - `func(args []interface{}) []interface{}` - 返回多个值

#### 关键方法

##### Register 方法
```go
func (s *Server) Register(id interface{}, f interface{})
```
- 注册函数到服务器
- 验证函数签名是否合法
- 防止重复注册

##### Go 方法（异步调用）
```go
func (s *Server) Go(id interface{}, args ...interface{})
```
- 异步调用函数，不等待结果
- 协程安全，可在多个 goroutine 中并发调用

##### Call0/Call1/CallN 方法（同步调用）
```go
func (s *Server) Call0(id interface{}, args ...interface{}) error
func (s *Server) Call1(id interface{}, args ...interface{}) (interface{}, error)
func (s *Server) CallN(id interface{}, args ...interface{}) ([]interface{}, error)
```
- 同步调用函数，等待结果返回
- 根据返回值类型区分三种调用方式
- 协程安全

### 2. Client（客户端）

#### 结构体定义
```go
type Client struct {
    s               *Server    // 关联的服务器
    chanSyncRet     chan *RetInfo   // 同步调用结果通道
    ChanAsynRet     chan *RetInfo   // 异步调用结果通道
    pendingAsynCall int            // 待处理的异步调用数量
}
```

#### 功能特点
- **单协程模型**：每个 Client 实例必须在单独的 goroutine 中使用（非协程安全）
- **双向通信**：支持同步和异步调用
- **流量控制**：限制待处理的异步调用数量

#### 关键方法

##### 同步调用
```go
func (c *Client) Call0(id interface{}, args ...interface{}) error
func (c *Client) Call1(id interface{}, args ...interface{}) (interface{}, error)
func (c *Client) CallN(id interface{}, args ...interface{}) ([]interface{}, error)
```

##### 异步调用
```go
func (c *Client) AsynCall(id interface{}, _args ...interface{})
```
- 最后一个参数必须是回调函数
- 支持三种回调函数签名：
  - `func(error)` - 对应无返回值的调用
  - `func(interface{}, error)` - 对应单个返回值的调用
  - `func([]interface{}, error)` - 对应多个返回值的调用

##### 回调处理
```go
func (c *Client) Cb(ri *RetInfo)
```
- 处理异步调用的结果
- 自动减少待处理调用计数

## 数据流结构

### CallInfo（调用信息）
```go
type CallInfo struct {
    f       interface{}     // 要执行的函数
    args    []interface{}   // 函数参数
    chanRet chan *RetInfo   // 返回结果通道
    cb      interface{}     // 回调函数
}
```

### RetInfo（返回信息）
```go
type RetInfo struct {
    ret interface{}     // 返回值
    err error           // 错误信息
    cb  interface{}     // 回调函数
}
```

## 执行流程

### 1. 同步调用流程

```
Client.CallX() -> Client.call() -> Server.ChanCall <- CallInfo
                                        ↓
                                    Server.exec() -> 执行函数
                                        ↓
                                    Server.ret() -> chanRet <- RetInfo
                                        ↓
                                Client接收结果并返回
```

### 2. 异步调用流程

```
Client.AsynCall() -> Client.asynCall() -> Server.ChanCall <- CallInfo
                                            ↓
                                        Server.exec() -> 执行函数
                                            ↓
                                        Server.ret() -> ChanAsynRet <- RetInfo
                                            ↓
                                    Client.Cb() -> 执行回调
```

## 错误处理

### panic 恢复机制
- 所有执行函数的地方都有 panic 恢复
- 发生 panic 时返回错误信息给调用方
- 支持堆栈跟踪（通过 conf.LenStackBuf 配置）

### 错误类型
- 函数未注册错误
- 函数签名不匹配错误
- 通道已满错误
- 服务器关闭错误

## 协程安全性

| 组件 | 协程安全 | 说明 |
|------|----------|------|
| Server.Register | 否 | 必须在调用 Go/Open 之前完成 |
| Server.Go | 是 | 可在多个 goroutine 中并发调用 |
| Server.CallX | 是 | 通过 Open 创建 Client 实现 |
| Client.CallX | 否 | 单个 Client 实例只能在单个 goroutine 中使用 |
| Client.AsynCall | 否 | 单个 Client 实例只能在单个 goroutine 中使用 |

## 使用示例

### 基本使用

```go
// 创建服务器
server := NewServer(100)

// 注册函数
server.Register("add", func(args []interface{}) interface{} {
    return args[0].(int) + args[1].(int)
})

// 在单独的 goroutine 中运行服务器
go func() {
    for ci := range server.ChanCall {
        server.Exec(ci)
    }
}()

// 同步调用
client := server.Open(0)
result, err := client.Call1("add", 1, 2)

// 异步调用
client.AsynCall("add", 1, 2, func(ret interface{}, err error) {
    if err == nil {
        fmt.Println("Result:", ret)
    }
})
```

## 设计特点

1. **简洁高效**：基于 Go channel 实现，避免了复杂的网络协议
2. **类型灵活**：使用 interface{} 作为参数和返回值类型，支持任意类型
3. **错误恢复**：完善的 panic 恢复机制，防止单个调用失败影响整个系统
4. **流量控制**：通过 channel 缓冲区和待处理调用计数实现流量控制
5. **回调支持**：异步调用支持回调函数，避免阻塞

## 注意事项

1. **单协程限制**：Server 和 Client 实例都必须在单独的 goroutine 中使用
2. **注册顺序**：所有函数必须在调用 Go/Open 之前注册完成
3. **内存管理**：异步调用数量不能超过 channel 缓冲区容量
4. **类型断言**：调用方需要确保参数和返回值的类型正确性