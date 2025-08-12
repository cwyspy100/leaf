# MongoDB 模块使用说明文档

## 概述

`mongodb.go` 提供了一个基于 mgo.v2 驱动的 MongoDB 连接池管理器，实现了线程安全的 MongoDB 会话管理、自动负载均衡和常用数据库操作封装。

## 核心组件

### 1. Session（会话）

#### 结构体定义
```go
type Session struct {
    *mgo.Session  // 嵌入 mgo 原生会话
    ref   int     // 引用计数
    index int     // 在堆中的索引位置
}
```

#### 功能特点
- **引用计数**：跟踪当前会话被使用的次数
- **索引管理**：支持在最小堆中的快速定位
- **嵌入扩展**：继承 mgo.Session 的所有方法

### 2. SessionHeap（会话堆）

#### 实现原理
```go
type SessionHeap []*Session
```

基于 Go 的 `container/heap` 实现的最小堆，用于管理会话池：

- **最小堆排序**：按引用计数 `ref` 升序排列
- **自动负载均衡**：优先使用引用次数最少的会话
- **动态调整**：通过 `heap.Fix` 在引用计数变化时重新排序

#### 堆操作方法
- `Len() int`：返回堆中元素数量
- `Less(i, j int) bool`：按引用计数比较
- `Swap(i, j int)`：交换元素位置并更新索引
- `Push(x interface{})`：添加新会话到堆尾
- `Pop() interface{}`：移除并返回堆顶元素

### 3. DialContext（连接上下文）

#### 结构体定义
```go
type DialContext struct {
    sync.Mutex        // 互斥锁保证线程安全
    sessions SessionHeap  // 会话池（最小堆实现）
}
```

#### 功能特点
- **线程安全**：所有操作都通过互斥锁保护
- **连接池管理**：维护一组可复用的 MongoDB 会话
- **自动故障恢复**：会话引用计数异常检测

## 核心功能详解

### 1. 连接建立

#### Dial（默认连接）
```go
func Dial(url string, sessionNum int) (*DialContext, error)
```
- **参数说明**：
  - `url`：MongoDB 连接字符串
  - `sessionNum`：连接池大小（会话数量）
- **默认配置**：10秒连接超时，5分钟操作超时

#### DialWithTimeout（自定义超时）
```go
func DialWithTimeout(url string, sessionNum int, dialTimeout time.Duration, timeout time.Duration) (*DialContext, error)
```
- **参数说明**：
  - `dialTimeout`：连接建立超时时间
  - `timeout`：读写操作超时时间
- **容错处理**：当 `sessionNum <= 0` 时自动重置为100

#### 初始化流程
1. 建立 MongoDB 主连接
2. 设置同步和套接字超时
3. 创建指定数量的会话副本
4. 初始化最小堆结构
5. 返回准备好的连接上下文

### 2. 会话管理

#### Ref（获取会话）
```go
func (c *DialContext) Ref() *Session
```
- **负载均衡**：总是返回引用计数最少的会话
- **自动刷新**：长时间未使用的会话会自动刷新
- **引用计数**：获取时自动增加引用计数
- **线程安全**：通过互斥锁保护

#### UnRef（释放会话）
```go
func (c *DialContext) UnRef(s *Session)
```
- **引用递减**：减少会话的引用计数
- **堆重排**：自动调整最小堆顺序
- **资源管理**：确保会话正确回收

#### 使用模式
```go
s := c.Ref()      // 获取会话
defer c.UnRef(s)  // 确保释放
// 使用 s 进行数据库操作
```

### 3. 计数器功能

#### EnsureCounter（确保计数器存在）
```go
func (c *DialContext) EnsureCounter(db string, collection string, id string) error
```
- **功能**：确保指定的计数器文档存在
- **幂等性**：重复调用不会报错（利用唯一索引冲突忽略）
- **文档结构**：
  ```json
  {
    "_id": "counter_id",
    "seq": 0
  }
  ```

#### NextSeq（获取下一个序列号）
```go
func (c *DialContext) NextSeq(db string, collection string, id string) (int, error)
```
- **原子操作**：使用 MongoDB 的 findAndModify 确保并发安全
- **自动递增**：每次调用序列号加1
- **返回值**：返回递增后的新序列号

### 4. 索引管理

#### EnsureIndex（普通索引）
```go
func (c *DialContext) EnsureIndex(db string, collection string, key []string) error
```
- **索引类型**：非唯一、稀疏索引
- **适用场景**：常规查询优化
- **参数**：
  - `key`：索引字段列表，如 `["name", "age"]`

#### EnsureUniqueIndex（唯一索引）
```go
func (c *DialContext) EnsureUniqueIndex(db string, collection string, key []string) error
```
- **索引类型**：唯一、稀疏索引
- **适用场景**：防止重复数据
- **约束**：确保指定字段组合的唯一性

## 使用示例

### 1. 基本连接
```go
package main

import (
    "log"
    "github.com/name5566/leaf/db/mongodb"
)

func main() {
    // 建立连接
    ctx, err := mongodb.Dial("mongodb://localhost:27017", 50)
    if err != nil {
        log.Fatal(err)
    }
    defer ctx.Close()

    // 使用连接...
}
```

### 2. 计数器使用
```go
// 确保计数器存在
err := ctx.EnsureCounter("mydb", "counters", "user_id")
if err != nil {
    log.Fatal(err)
}

// 获取下一个用户ID
userID, err := ctx.NextSeq("mydb", "counters", "user_id")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("New user ID: %d\n", userID)
```

### 3. 索引创建
```go
// 创建普通索引
err := ctx.EnsureIndex("mydb", "users", []string{"email", "status"})
if err != nil {
    log.Fatal(err)
}

// 创建唯一索引
err = ctx.EnsureUniqueIndex("mydb", "users", []string{"username"})
if err != nil {
    log.Fatal(err)
}
```

### 4. 数据操作
```go
// 获取会话
s := ctx.Ref()
defer ctx.UnRef(s)

// 插入文档
coll := s.DB("mydb").C("users")
err := coll.Insert(bson.M{
    "_id":      userID,
    "username": "john_doe",
    "email":    "john@example.com",
    "status":   "active",
})
```

## 性能特点

### 1. 连接池优势
- **复用连接**：避免频繁创建/销毁连接的开销
- **负载均衡**：通过最小堆实现请求均匀分布
- **资源控制**：限制最大并发连接数

### 2. 线程安全
- **锁粒度优化**：使用细粒度锁保护会话池
- **无锁操作**：会话本身的操作无需加锁
- **并发友好**：支持高并发环境下的安全使用

### 3. 故障检测
- **引用计数监控**：Close 时检测未正确释放的会话
- **日志记录**：异常情况下的详细错误日志
- **资源泄露防护**：自动检测会话泄露

## 最佳实践

### 1. 连接管理
```go
// 推荐：全局连接池
var mongoCtx *mongodb.DialContext

func init() {
    var err error
    mongoCtx, err = mongodb.Dial("mongodb://localhost:27017", 100)
    if err != nil {
        log.Fatal(err)
    }
}

func cleanup() {
    if mongoCtx != nil {
        mongoCtx.Close()
    }
}
```

### 2. 错误处理
```go
// 始终检查错误
if err := ctx.EnsureCounter("db", "counters", "order"); err != nil {
    // 处理错误：重试、记录日志或返回错误
    log.Printf("Failed to ensure counter: %v", err)
    return err
}
```

### 3. 资源管理
```go
// 使用 defer 确保资源释放
func createUser(ctx *mongodb.DialContext, userData interface{}) error {
    s := ctx.Ref()
    defer ctx.UnRef(s)
    
    userID, err := ctx.NextSeq("mydb", "counters", "user_id")
    if err != nil {
        return err
    }
    
    return s.DB("mydb").C("users").Insert(userData)
}
```

## 注意事项

1. **连接数配置**：根据实际负载合理设置连接池大小
2. **超时设置**：为长时间运行的操作设置适当的超时时间
3. **错误处理**：始终检查并适当处理返回的错误
4. **资源释放**：务必使用 `defer` 确保会话正确释放
5. **索引策略**：在大量数据操作前预先创建必要的索引

## 总结

该 MongoDB 模块提供了一个生产就绪的连接池解决方案，具有以下特点：

- **简单易用**：封装了复杂的数据库连接管理
- **高性能**：最小堆实现的负载均衡算法
- **线程安全**：完善的并发控制机制
- **功能完备**：涵盖常用的数据库操作需求
- **可扩展**：基于 mgo.v2 驱动，支持所有 MongoDB 特性

适合在游戏服务器、Web 应用等需要高并发 MongoDB 访问的场景中使用。