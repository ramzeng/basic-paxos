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
	acceptedMessages map[int]Message
}

func (l *Learner) Learn(message *Message, reply *Reply) error {
	m := l.acceptedMessages[message.From]

	// 如果提案的编号大于已接受的提案编号，则接受该提案
	if m.ProposalNumber < message.ProposalNumber {
		l.acceptedMessages[message.From] = *message
		reply.OK = true
	} else {
		reply.OK = false
	}

	return nil
}

func (l *Learner) Chosen() interface{} {
	acceptorsCount := make(map[int]int)
	acceptedMessages := make(map[int]Message)

	for _, message := range l.acceptedMessages {
		// 如果提案的编号不为 0，则接受该提案
		if message.ProposalNumber != 0 {
			acceptorsCount[message.ProposalNumber]++
			acceptedMessages[message.ProposalNumber] = message
		}
	}

	for proposalNumber, count := range acceptorsCount {
		if count > l.halfAcceptedMessagesCount() {
			return acceptedMessages[proposalNumber].ProposalValue
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

func (l *Learner) halfAcceptedMessagesCount() int {
	return len(l.acceptedMessages) / 2
}

func NewLearner(id int, acceptorIds []int) *Learner {
	learner := &Learner{
		id:               id,
		acceptedMessages: make(map[int]Message),
	}

	for _, acceptorId := range acceptorIds {
		learner.acceptedMessages[acceptorId] = Message{}
	}

	learner.Serve(id)

	return learner
}
