package main

import "net/rpc"

type Message struct {
	// 提案编号
	Number int
	// 提案值
	Value interface{}
	// 发送者 ID
	From int
	// 接收者 ID
	To int
}

type Reply struct {
	OK     bool
	Number int
	Value  interface{}
}

func call(address string, name string, arguments interface{}, reply interface{}) error {
	dial, err := rpc.Dial("tcp", address)

	if err != nil {
		return err
	}

	defer func() {
		_ = dial.Close()
	}()

	return dial.Call(name, arguments, reply)
}
