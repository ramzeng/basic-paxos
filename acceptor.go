package main

import (
	"fmt"
	"net"
	"net/rpc"
)

// Acceptor 接受者
type Acceptor struct {
	listener net.Listener
	// 节点 ID
	id int
	// 承诺的提案编号，如果为 0，则表示没有收到过任何 Prepare 消息
	proposalProposal int
	// 已接受的提案编号，如果为 0，表示没有接受任何提案
	acceptedProposalNumber int
	// 已接受的提案值
	acceptedProposalValue interface{}
	// 学习者 ID 列表
	learners []int
}

func (a *Acceptor) Prepare(message *Message, reply *Reply) error {
	// 如果提案编号大于当前承诺的提案编号，则承诺提案编号
	if message.ProposalNumber > a.proposalProposal {
		a.proposalProposal = message.ProposalNumber

		// 返回已接受的提案编号和提案值
		reply.ProposalNumber = a.acceptedProposalNumber
		reply.ProposalValue = a.acceptedProposalValue
		reply.OK = true
	} else {
		// 否则，拒绝提案
		reply.OK = false
	}

	return nil
}

func (a *Acceptor) Accept(message *Message, reply *Reply) error {
	// 如果提案编号大于等于当前承诺的提案编号，则接受提案
	if message.ProposalNumber >= a.proposalProposal {
		a.proposalProposal = message.ProposalNumber

		// 记录已接受的提案编号和提案值
		a.acceptedProposalNumber = message.ProposalNumber
		a.acceptedProposalValue = message.ProposalValue

		reply.OK = true

		// 向所有学习者发送提案
		for _, learner := range a.learners {
			go func(learner int) {
				message.From = a.id
				message.To = learner

				if err := call(
					fmt.Sprintf("127.0.0.1:%d", learner),
					"Learner.Learn",
					message,
					&Reply{},
				); err != nil {
					return
				}
			}(learner)
		}
	} else {
		reply.OK = false
	}

	return nil
}

func (a *Acceptor) Serve() {
	server := rpc.NewServer()

	if err := server.Register(a); err != nil {
		panic(err)
	}

	if listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.id)); err != nil {
		panic(err)
	} else {
		a.listener = listen
	}

	go func() {
		for {
			connection, err := a.listener.Accept()

			if err != nil {
				continue
			}

			go server.ServeConn(connection)
		}
	}()
}

func (a *Acceptor) Close() {
	_ = a.listener.Close()
}

func NewAcceptor(id int, learners []int) *Acceptor {
	acceptor := &Acceptor{
		id:       id,
		learners: learners,
	}

	acceptor.Serve()

	return acceptor
}
