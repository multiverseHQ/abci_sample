package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tendermint/tendermint/rpc/client"
)

type node struct {
	Address string
	Key     string

	cli *client.HTTP
}

type nodeInfo struct {
	Key         string
	VotingPower int64
	Reward      uint64
	TxCount     uint64
}

type nodeManager struct {
	Keys          []string
	lastNodeIdx   int
	nodeByAddress map[string]*node
	nodeByKey     map[string]*node
}

func newNodeByAddress(address string) (*node, error) {
	cli := client.NewHTTP(address, "/websocket")

	status, err := cli.Status()
	if err != nil {
		return nil, err
	}

	res := &node{
		Address: address,
		Key:     hex.EncodeToString(status.PubKey.Address()),
		cli:     cli,
	}

	return res, res.setMyKey()
}

var totalTx = -1

func (n *node) makeTx(fail bool) error {
	totalTx++
	amount := totalTx
	if fail == true {
		amount = amount - 1
	}

	args := map[string]interface{}{
		"Amount":             amount,
		"reward_destination": n.Key,
	}

	raw, _ := json.Marshal(args)
	encoded := base64.StdEncoding.EncodeToString(raw)
	res, err := n.cli.BroadcastTxCommit([]byte(encoded))
	if res.CheckTx.IsErr() {
		return fmt.Errorf("Invalid transaction: %s", res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return fmt.Errorf("Cannot deliver transaction: %s", res.DeliverTx.Error())
	}
	return err
}

func (n *node) setMyKey() error {
	_, err := n.cli.ABCIQuery("set_key", []byte(n.Key), false)
	return err
}

func (n *node) getTx() (uint64, error) {
	txRes, err := n.cli.ABCIQuery("tx", nil, false)
	if err != nil {
		return 0, err
	}
	log.Printf("Got %d:%v", len(txRes.Value), txRes.Value)
	return binary.BigEndian.Uint64(txRes.Value), nil
}

func (n *node) getRewards() (map[string]uint64, error) {
	rewardRes, err := n.cli.ABCIQuery("rewards", nil, false)
	if err != nil {
		return nil, err
	}
	res := make(map[string]uint64)
	return res, json.Unmarshal(rewardRes.Value, &res)
}

func (n *nodeManager) getNodeInfo() ([]*nodeInfo, error) {
	defer func() { n.lastNodeIdx = (n.lastNodeIdx + 1) % len(n.Keys) }()

	infos := make(map[string]*nodeInfo)

	nn := n.nodeByKey[n.Keys[n.lastNodeIdx]]

	txNumber, err := nn.getTx()
	if err != nil {
		return nil, err
	}
	rewards, err := nn.getRewards()
	if err != nil {
		return nil, err
	}

	for k, r := range rewards {
		ni := &nodeInfo{
			Key:         k,
			TxCount:     txNumber,
			Reward:      r,
			VotingPower: 0,
		}
		infos[k] = ni
	}

	validators, err := nn.cli.Validators()
	if err != nil {
		return nil, err
	}
	for _, v := range validators.Validators {
		k := string(v.PubKey.Address())
		if ni, ok := infos[k]; ok == true {
			ni.VotingPower = v.VotingPower
		} else {
			infos[k] = &nodeInfo{
				Key:         k,
				TxCount:     txNumber,
				Reward:      0,
				VotingPower: v.VotingPower,
			}
		}
	}

	res := make([]*nodeInfo, 0, len(infos))
	for _, i := range infos {
		res = append(res, i)
	}

	return res, nil
}
