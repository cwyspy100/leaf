package chanrpc

import (
	"fmt"
	"testing"
)

func TestClient_AsynCall(t *testing.T) {
	// 创建服务器
	server := NewServer(100)

	// 注册函数
	server.Register("add", func(args []interface{}) interface{} {
		// 当从 interface{} 中取出值时，需要通过类型断言（如 v.(int)）还原为原始类型，否则无法直接使用
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

	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)

	/*// 异步调用
	client.AsynCall("add", 1, 2, func(ret interface{}, err error) {
		if err == nil {
			fmt.Println("Result:", ret)
		}
	})*/
}
