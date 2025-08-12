package main

import (
	"fmt"
	"github.com/name5566/leaf/util"
)

type User struct {
	Name    string
	Age     int
	Hobbies []string `deepcopy:"-"` // 标记为跳过拷贝
	Info    map[string]interface{}
}

func main() {
	src := User{
		Name:    "Alice",
		Age:     30,
		Hobbies: []string{"reading", "sports"},
		Info: map[string]interface{}{
			"city": "Beijing",
		},
	}

	// 准备目标变量（需要与src同类型）
	var dst User
	fmt.Printf("Before copy - src.Info: %+v\n", src.Info)
	fmt.Printf("Before copy - src.Info type: %T\n", src.Info["city"])
	// 反射获取src和dst的Value
	util.DeepCopy(&src, &dst)
	fmt.Printf("After copy - dst.Info: %+v\n", dst.Info)
	//util.deepCopy(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))

	// 修改dst不会影响src
	dst.Name = "Bob"
	if dst.Info == nil {
		dst.Info = make(map[string]interface{})
	}
	dst.Info["city"] = "Shanghai"
	fmt.Printf("src: %+v\n", src)
	fmt.Printf("dst: %+v\n", dst)
	fmt.Printf("src.Name: %s\n", src.Name)
	fmt.Printf("src.Info[\"city\"]: %v\n", src.Info["city"])
	fmt.Printf("dst.Name: %s\n", dst.Name)
	fmt.Printf("dst.Info[\"city\"]: %v\n", dst.Info["city"])
	fmt.Printf("dst.Hobbies: %v\n", dst.Hobbies)

	// Verify deep copy worked correctly
	fmt.Printf("Are src and dst independent? src.Info[\"city\"]=%v, dst.Info[\"city\"]=%v\n", src.Info["city"], dst.Info["city"])
}
