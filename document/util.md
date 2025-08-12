# Util 模块实现原理分析

## 概述

Util 模块提供了通用的工具函数集合，包含以下核心功能：
- **深度拷贝**：支持任意类型的深度复制
- **并发安全Map**：基于sync.RWMutex的线程安全映射
- **随机数工具**：概率分组、区间随机等实用函数
- **信号量**：基于channel的简单信号量实现

## 核心组件

### 1. DeepCopy - 深度拷贝

#### 实现原理

基于反射实现任意类型的深度拷贝，支持以下类型：
- 基本类型（int, string, bool等）
- 复合类型（struct, slice, map, array）
- 指针类型（*T, interface{}）

#### 核心算法

```go
// 递归深度拷贝算法
func deepCopy(dst, src reflect.Value) {
    switch src.Kind() {
    case reflect.Interface:
        // 处理接口类型
        value := src.Elem()
        newValue := reflect.New(value.Type()).Elem()
        deepCopy(newValue, value)
        dst.Set(newValue)
        
    case reflect.Ptr:
        // 处理指针类型
        dst.Set(reflect.New(value.Type()))
        deepCopy(dst.Elem(), value)
        
    case reflect.Map:
        // 处理Map类型：创建新Map，递归拷贝键值对
        dst.Set(reflect.MakeMap(src.Type()))
        for key, value := range keys {
            newValue := reflect.New(value.Type()).Elem()
            deepCopy(newValue, value)
            dst.SetMapIndex(key, newValue)
        }
        
    case reflect.Slice:
        // 处理切片类型：创建新切片，逐个元素拷贝
        dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
        for i := 0; i < src.Len(); i++ {
            deepCopy(dst.Index(i), src.Index(i))
        }
        
    case reflect.Struct:
        // 处理结构体：检查deepcopy标签，逐个字段拷贝
        for i := 0; i < src.NumField(); i++ {
            if tag.Get("deepcopy") != "-" {
                deepCopy(dst.Field(i), src.Field(i))
            }
        }
        
    default:
        // 基本类型直接赋值
        dst.Set(src)
    }
}
```

#### 使用方式

```go
// 方式1：指针拷贝
var dst []int
util.DeepCopy(&dst, &src)

// 方式2：克隆新对象
dst := util.DeepClone(src).([]int)
```

### 2. Map - 并发安全映射

#### 实现原理

基于 `sync.RWMutex` 实现的线程安全映射，提供以下特性：

#### 核心设计

```go
type Map struct {
    sync.RWMutex
    m map[interface{}]interface{}
}
```

#### 功能特性

1. **延迟初始化**：首次使用时自动初始化底层map
2. **读写分离**：使用RWMutex实现读写分离，提高并发性能
3. **原子操作**：提供TestAndSet等原子操作
4. **遍历支持**：支持安全遍历和范围操作

#### 性能特点

- **读多写少优化**：多个goroutine可同时读取
- **锁粒度精细**：每个Map实例独立锁，无全局锁竞争
- **内存安全**：避免并发访问时的panic

### 3. Rand - 随机数工具

#### 实现原理

基于 `math/rand` 的高级随机数封装，提供以下功能：

#### 概率分组算法

```go
// RandGroup 实现加权随机选择
func RandGroup(p ...uint32) int {
    // 计算累计概率分布
    r := make([]uint32, len(p))
    for i := 0; i < len(p); i++ {
        if i == 0 {
            r[0] = p[0]
        } else {
            r[i] = r[i-1] + p[i]
        }
    }
    
    // 随机选择区间
    rn := uint32(rand.Int63n(int64(r[len(r)-1])))
    for i := 0; i < len(r); i++ {
        if rn < r[i] {
            return i
        }
    }
}
```

#### 无重复随机算法

```go
// RandIntervalN 实现区间内无重复随机数
func RandIntervalN(b1, b2 int32, n uint32) []int32 {
    // 使用Fisher-Yates洗牌算法变种
    m := make(map[int32]int32) // 位置映射
    for i := uint32(0); i < n; i++ {
        v := int32(rand.Int63n(l) + min)
        
        // 处理冲突映射
        if mv, ok := m[v]; ok {
            r[i] = mv
        } else {
            r[i] = v
        }
        
        // 更新映射
        lv := int32(l - 1 + min)
        if v != lv {
            m[v] = m[lv]
        }
    }
}
```

### 4. Semaphore - 信号量

#### 实现原理

基于 buffered channel 实现的简单信号量：

```go
type Semaphore chan struct{}

// 创建信号量
func MakeSemaphore(n int) Semaphore {
    return make(Semaphore, n)
}

// 获取信号量
func (s Semaphore) Acquire() {
    s <- struct{}{}
}

// 释放信号量
func (s Semaphore) Release() {
    <-s
}
```

#### 特点
- **简洁实现**：仅3行代码实现完整功能
- **阻塞语义**：Acquire在信号量为0时阻塞
- **无死锁**：Release必须在Acquire后调用

## 优点分析

### 1. 深度拷贝
- **类型通用**：支持任意Go类型，无需注册
- **性能优化**：跳过不可导出字段，避免无效拷贝
- **内存安全**：完全独立的拷贝，无引用共享
- **标签支持**：通过`deepcopy:"-"`控制拷贝行为

### 2. 并发安全Map
- **读写优化**：RWMutex在读多写少场景性能优异
- **API丰富**：提供Unsafe系列方法供内部优化使用
- **延迟初始化**：避免不必要的内存分配
- **原子操作**：TestAndSet避免竞态条件

### 3. 随机数工具
- **算法高效**：概率分组使用O(n)预计算，O(log n)查询
- **无重复算法**：Fisher-Yates变种，时间复杂度O(n)
- **边界处理**：自动处理边界条件和参数验证
- **线程安全**：使用全局随机源，自动初始化

### 4. 信号量
- **实现简洁**：channel天然支持阻塞和唤醒
- **零开销**：无额外内存分配
- **语义清晰**：符合信号量标准定义

## 缺点分析

### 1. 深度拷贝
- **性能开销**：反射调用开销较大，比手动拷贝慢10-100倍
- **类型限制**：无法处理函数类型、channel类型
- **循环引用**：可能导致无限递归和栈溢出
- **接口限制**：需要指针参数，使用不够直观

### 2. 并发Map
- **锁竞争**：高并发写场景性能下降明显
- **GC压力**：interface{}类型增加GC负担
- **无类型安全**：运行时类型检查，编译时无法发现类型错误
- **迭代器缺失**：不支持类似sync.Map的Range方法

### 3. 随机数工具
- **全局锁**：使用全局rand.Source，高并发存在锁竞争
- **精度限制**：32位整数范围限制
- **内存分配**：RandIntervalN需要额外内存分配
- **算法复杂度**：无重复算法在n接近区间大小时性能下降

### 4. 信号量
- **功能简单**：仅支持基本Acquire/Release
- **无超时机制**：不支持带超时的Acquire
- **容量固定**：创建后无法调整容量
- **无监控**：无法获取当前可用信号量数量

## 性能对比

| 功能 | 本实现 | 标准库 | 性能对比 |
|------|--------|--------|----------|
| 深度拷贝 | 反射实现 | 无 | 较慢但通用 |
| 并发Map | RWMutex | sync.Map | 读多写少更优 |
| 随机数 | math/rand封装 | math/rand | 功能更丰富 |
| 信号量 | channel实现 | golang.org/x/sync/semaphore | 更简洁 |

## 使用示例

### 深度拷贝
```go
// 结构体拷贝
type User struct {
    Name string
    Age  int
    Data map[string]interface{}
}

src := &User{Name: "Alice", Age: 30, Data: map[string]interface{}{"key": "value"}}
var dst User
util.DeepCopy(&dst, src)

// 切片拷贝
src := []int{1, 2, 3}
dst := util.DeepClone(src).([]int)
```

### 并发Map
```go
m := new(util.Map)

// 基本操作
m.Set("key", "value")
v := m.Get("key")
m.Del("key")

// 并发安全遍历
m.RLockRange(func(k, v interface{}) {
    fmt.Printf("%v: %v\n", k, v)
})

// 原子操作
if old := m.TestAndSet("key", "new"); old == nil {
    fmt.Println("new key set")
}
```

### 随机数工具
```go
// 概率分组
result := util.RandGroup(10, 30, 60) // 返回0,1,2的概率分别为10%,30%,60%

// 区间随机
num := util.RandInterval(1, 100) // 1-100之间的随机数

// 无重复随机数
nums := util.RandIntervalN(1, 100, 10) // 1-100之间10个不重复的随机数
```

### 信号量
```go
sem := util.MakeSemaphore(10)

// 使用示例
sem.Acquire()
defer sem.Release()

// 并发控制
go func() {
    sem.Acquire()
    defer sem.Release()
    // 执行受控操作
}()
```

## 适用场景

1. **游戏服务器**：玩家数据深度拷贝、并发数据管理
2. **Web服务**：配置热加载、缓存管理
3. **数据处理**：ETL过程中的数据复制
4. **并发控制**：连接池、资源池管理
5. **测试工具**：随机数据生成、并发测试