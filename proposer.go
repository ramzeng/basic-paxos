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

func (p *Proposer) Propose(v interface{}) interface{} {
	p.round++
	p.number = p.proposalNumber()

	// Prepare 阶段
	prepareCount := 0
	maxNumber := 0

	for _, acceptor := range p.acceptors {
		message := Message{
			Number: p.number,
			From:   p.id,
			To:     acceptor,
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
			prepareCount++
			// 如果收到的提案编号比当前的大，就更新当前的提案编号和提案值
			if reply.Number > maxNumber {
				maxNumber = reply.Number
				v = reply.Value
			}
		}

		if prepareCount > len(p.acceptors)/2 {
			break
		}
	}

	// Accept 阶段
	acceptCount := 0
	if prepareCount > len(p.acceptors)/2 {
		for _, acceptor := range p.acceptors {
			message := Message{
				Number: p.number,
				Value:  v,
				From:   p.id,
				To:     acceptor,
			}

			reply := &Reply{}

			if err := call(fmt.Sprintf("127.0.0.1:%d", acceptor), "Acceptor.Accept", message, reply); err != nil {
				continue
			}

			if reply.OK {
				acceptCount++
			}
		}
	}

	if acceptCount > len(p.acceptors)/2 {
		return v
	}

	return nil
}

func (p *Proposer) proposalNumber() int {
	// 提案编号 = (轮次,节点 ID)
	return p.round<<16 | p.id
}
