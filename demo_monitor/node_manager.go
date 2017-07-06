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

func (nm *nodeManager) commitTx(failit bool) error {
	defer func() { nm.curNodeIdx = (nm.curNodeIdx + 1) % len(nm.Keys) }()
	n := nm.nodeByKey[nm.Keys[nm.curNodeIdx]]
	tx, err := n.getTx()
	if err != nil {
		return err
	}

	if failit == true {
		tx--
	}

	return n.makeTx(tx)
}

func (nm *nodeManager) fetchStatus() ([]*nodeInfo, error) {
	defer func() { nm.curNodeIdx = (nm.curNodeIdx + 1) % len(nm.Keys) }()
	n := nm.nodeByKey[nm.Keys[nm.curNodeIdx]]

	infos, err := n.getAllNodeInfo()
	if err != nil {
		return infos, err
	}

	// CHECK FOR NEW NODES !!!
	hasNewNode := false
	for _, i := range infos {
		if _, ok := nm.nodeByKey[i.Key]; ok == false {
			hasNewNode = true
		}
	}

	if hasNewNode == false {
		return infos, nil
	}

	// fetch the new node !!
	peers, err := n.getPeers()
	if err != nil {
		return infos, err
	}

	for _, p := range peers {
		if _, ok := nm.nodeByAddress[p]; ok == true {
			continue
		}
		nn, err := newNodeByAddress(p)
		if err != nil {
			return nil, err
		}
		nm.Keys = append(nm.Keys, nn.Key)
		nm.nodeByAddress[p] = nn
		nm.nodeByKey[nn.Key] = nn
	}

	return infos, nil
}
