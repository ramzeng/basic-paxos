package main

import "fmt"

// Proposer 提议者
type Proposer struct {
	// 节点 ID
	id int
	// 最大轮次
	round int
	// 提案编号
	number int
	// 接受者 ID 列表
	acceptors []int
}

func (p *Proposer) Propose(value interface{}) interface{} {
	p.round++
	p.number = p.proposalNumber()

	// Prepare 阶段
	preparedAcceptorsCount := 0
	maxProposerNumber := 0

	for _, acceptor := range p.acceptors {
		message := Message{
			ProposalNumber: p.number,
			From:           p.id,
			To:             acceptor,
		}

		reply := &Reply{}

		if err := call(
			fmt.Sprintf("127.0.0.1:%d", acceptor),
			"Acceptor.Prepare",
			message,
			reply,
		); err != nil {
			continue
		}

		if reply.OK {
			preparedAcceptorsCount++
			// 如果收到的提案编号比当前的大，就更新当前的提案编号和提案值
			if reply.ProposalNumber > maxProposerNumber {
				maxProposerNumber = reply.ProposalNumber
				value = reply.ProposalValue
			}
		}

		if preparedAcceptorsCount > p.halfAcceptorsCount() {
			break
		}
	}

	// Accept 阶段
	acceptedAcceptorsCount := 0

	if preparedAcceptorsCount > p.halfAcceptorsCount() {
		for _, acceptor := range p.acceptors {
			message := Message{
				ProposalNumber: p.number,
				ProposalValue:  value,
				From:           p.id,
				To:             acceptor,
			}

			reply := &Reply{}

			if err := call(
				fmt.Sprintf("127.0.0.1:%d", acceptor),
				"Acceptor.Accept",
				message,
				reply,
			); err != nil {
				continue
			}

			if reply.OK {
				acceptedAcceptorsCount++
			}
		}
	}

	if acceptedAcceptorsCount > p.halfAcceptorsCount() {
		return value
	}

	return nil
}

func (p *Proposer) proposalNumber() int {
	// 提案编号 = (轮次,节点 ID)
	return p.round<<16 | p.id
}

func (p *Proposer) halfAcceptorsCount() int {
	return len(p.acceptors) / 2
}
