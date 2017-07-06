package main

import "sort"

type nodeManager struct {
	Keys          []string
	curNodeIdx    int
	nodeByAddress map[string]*node
	nodeByKey     map[string]*node
}

func newNodeManager(firstNodeAddress string) (*nodeManager, error) {
	firstNode, err := newNodeByAddress(firstNodeAddress)
	if err != nil {
		return nil, err
	}

	res := &nodeManager{
		curNodeIdx: 0,
		nodeByAddress: map[string]*node{
			firstNode.Address: firstNode,
		},
		nodeByKey: map[string]*node{
			firstNode.Key: firstNode,
		},
		Keys: []string{firstNode.Key},
	}

	peers, err := firstNode.getPeers()
	if err != nil {
		return nil, err
	}

	for _, p := range peers {
		n, err := newNodeByAddress(p)
		if err != nil {
			return res, err
		}
		res.Keys = append(res.Keys, n.Key)
		res.nodeByAddress[n.Address] = n
		res.nodeByKey[n.Key] = n
	}
	//sort by key
	sort.Strings(res.Keys)

	return res, nil
}
