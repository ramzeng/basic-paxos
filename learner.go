package main

import (
	"fmt"
	"net"
	"net/rpc"
)

type Learner struct {
	listener net.Listener
	// 节点 ID
	id int
	// 已接受的提案
	acceptedMessage map[int]Message
}

func (l *Learner) Learn(message *Message, reply *Reply) error {
	m := l.acceptedMessage[message.From]

	// 如果提案的编号大于已接受的提案编号，则接受该提案
	if m.Number < message.Number {
		l.acceptedMessage[message.From] = *message
		reply.OK = true
	} else {
		reply.OK = false
	}

	return nil
}

func (l *Learner) Chosen() interface{} {
	acceptedCounts := make(map[int]int)
	acceptedMessages := make(map[int]Message)

	for _, message := range l.acceptedMessage {
		// 如果提案的编号不为 0，则接受该提案
		if message.Number != 0 {
			acceptedCounts[message.Number]++
			acceptedMessages[message.Number] = message
		}
	}

	for n, count := range acceptedCounts {
		if count > len(l.acceptedMessage)/2 {
			return acceptedMessages[n].Value
		}
	}

	return nil
}

func (l *Learner) Serve(id int) {
	server := rpc.NewServer()

	if err := server.Register(l); err != nil {
		panic(err)
	}

	if listen, err := net.Listen("tcp", fmt.Sprintf(":%d", id)); err != nil {
		panic(err)
	} else {
		l.listener = listen
	}

	go func() {
		for {
			connection, err := l.listener.Accept()

			if err != nil {
				continue
			}

			go server.ServeConn(connection)
		}
	}()
}

func (l *Learner) Close() {
	_ = l.listener.Close()
}

func NewLearner(id int, acceptorIds []int) *Learner {
	learner := &Learner{
		id:              id,
		acceptedMessage: make(map[int]Message),
	}

	for _, acceptorId := range acceptorIds {
		learner.acceptedMessage[acceptorId] = Message{}
	}

	learner.Serve(id)

	return learner
}
