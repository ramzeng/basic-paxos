package main

import "testing"

func start(acceptorIds []int, learnerIds []int) ([]*Acceptor, []*Learner) {
	acceptors := make([]*Acceptor, len(acceptorIds))

	for i, acceptorId := range acceptorIds {
		acceptors[i] = NewAcceptor(acceptorId, learnerIds)
	}

	learners := make([]*Learner, len(learnerIds))

	for i, learnerId := range learnerIds {
		learners[i] = NewLearner(learnerId, acceptorIds)
	}

	return acceptors, learners
}

func cleanup(acceptors []*Acceptor, learners []*Learner) {
	for _, acceptor := range acceptors {
		acceptor.Close()
	}

	for _, learner := range learners {
		learner.Close()
	}
}

func TestSingleProposer(t *testing.T) {
	acceptorIds := []int{1001, 1002, 1003}
	learnerIds := []int{2001}

	acceptors, learners := start(acceptorIds, learnerIds)

	defer cleanup(acceptors, learners)

	p := &Proposer{
		id:        1,
		acceptors: acceptorIds,
	}

	value := p.Propose("Hello World!")

	if value != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %s", value)
	}

	for _, learner := range learners {
		if learner.Chosen() != "Hello World!" {
			t.Errorf("Expected 'Hello World!', got %s", learner.Chosen())
		}
	}
}

func TestMultipleProposers(t *testing.T) {
	acceptorIds := []int{1001, 1002, 1003}
	learnerIds := []int{2001}

	acceptors, learners := start(acceptorIds, learnerIds)

	defer cleanup(acceptors, learners)

	p1 := &Proposer{
		id:        1,
		acceptors: acceptorIds,
	}

	v1 := p1.Propose("Hello World!")

	p2 := &Proposer{
		id:        2,
		acceptors: acceptorIds,
	}

	v2 := p2.Propose("Hello Paxos!")

	if v1 != v2 {
		t.Errorf("value1 = %s, value2 = %s", v1, v2)
	}

	for _, learner := range learners {
		if learner.Chosen() != v1 {
			t.Errorf("Expected %s, got %s", v1, learner.Chosen())
		}
	}
}
