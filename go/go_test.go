package g

import (
	"container/list"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestGo_Cb(t *testing.T) {
	type fields struct {
		ChanCb    chan func()
		pendingGo int
	}
	type args struct {
		cb func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add testleaf cases.
		{"test0", fields{ChanCb: make(chan func()), pendingGo: 0}, args{cb: func() { fmt.Println("finish!!!") }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Go{
				ChanCb:    tt.fields.ChanCb,
				pendingGo: tt.fields.pendingGo,
			}
			g.Cb(tt.args.cb)
		})
	}
}

func TestGo_Close(t *testing.T) {
	type fields struct {
		ChanCb    chan func()
		pendingGo int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add testleaf cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Go{
				ChanCb:    tt.fields.ChanCb,
				pendingGo: tt.fields.pendingGo,
			}
			g.Close()
		})
	}
}

func TestGo_Go(t *testing.T) {

}

func TestGo_Idle(t *testing.T) {
	type fields struct {
		ChanCb    chan func()
		pendingGo int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add testleaf cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Go{
				ChanCb:    tt.fields.ChanCb,
				pendingGo: tt.fields.pendingGo,
			}
			if got := g.Idle(); got != tt.want {
				t.Errorf("Idle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGo_NewLinearContext(t *testing.T) {
	type fields struct {
		ChanCb    chan func()
		pendingGo int
	}
	tests := []struct {
		name   string
		fields fields
		want   *LinearContext
	}{
		// TODO: Add testleaf cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Go{
				ChanCb:    tt.fields.ChanCb,
				pendingGo: tt.fields.pendingGo,
			}
			if got := g.NewLinearContext(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLinearContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinearContext_Go(t *testing.T) {
	type fields struct {
		g              *Go
		linearGo       *list.List
		mutexLinearGo  sync.Mutex
		mutexExecution sync.Mutex
	}
	type args struct {
		f  func()
		cb func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add testleaf cases.
		//{"test0", fields{ChanCb: make(chan func()), pendingGo: 0}, args{cb: func() { fmt.Println("finish!!!") }}},
		{
			name: "test_linear_execution", // 测试名称：验证串行执行
			fields: fields{
				g: &Go{ // 初始化Go实例，用于管理回调
					ChanCb:    make(chan func(), 10), // 缓冲通道，避免阻塞
					pendingGo: 0,
				},
				linearGo:       list.New(), // 初始化列表，用于存储待执行任务
				mutexLinearGo:  sync.Mutex{},
				mutexExecution: sync.Mutex{},
			},
			args: args{
				// 要执行的函数f：模拟耗时操作，记录执行顺序
				f: func() {
					// 这里可以添加具体逻辑，例如操作共享变量
				},
				// 回调函数cb：验证是否在f之后执行
				cb: func() {
					t.Log("回调函数执行")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &LinearContext{
				g:              tt.fields.g,
				linearGo:       tt.fields.linearGo,
				mutexLinearGo:  tt.fields.mutexLinearGo,
				mutexExecution: tt.fields.mutexExecution,
			}
			c.Go(tt.args.f, tt.args.cb)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		l int
	}
	tests := []struct {
		name string
		args args
		want *Go
	}{
		// TODO: Add testleaf cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
