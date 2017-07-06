package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tendermint/go-crypto"
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

func pubKeyToString(pk crypto.PubKey) string {
	return hex.EncodeToString(pk.Address())
}

func newNodeByAddress(address string) (*node, error) {
	cli := client.NewHTTP(address, "/websocket")

	status, err := cli.Status()
	if err != nil {
		return nil, err
	}

	res := &node{
		Address: address,
		Key:     pubKeyToString(status.PubKey),
		cli:     cli,
	}

	return res, res.setMyKey()
}

func (n *node) makeTx(nonce uint64) error {

	args := map[string]interface{}{
		"Amount":             nonce,
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

type byKey []*nodeInfo

func (l byKey) Len() int           { return len(l) }
func (l byKey) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l byKey) Less(i, j int) bool { return l[i].Key < l[j].Key }

func (n *node) getAllNodeInfo() ([]*nodeInfo, error) {
	infos := make(map[string]*nodeInfo)

	txNumber, err := n.getTx()
	if err != nil {
		return nil, err
	}

	rewards, err := n.getRewards()
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

	validators, err := n.cli.Validators()
	if err != nil {
		return nil, err
	}
	for _, v := range validators.Validators {
		k := pubKeyToString(v.PubKey)
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

	//we sort by key
	sort.Sort(byKey(res))
	return res, nil
}

var portRx = regexp.MustCompile(`:[0-9]+\z`)

func (n *node) getPeers() ([]string, error) {
	infos, err := n.cli.NetInfo()
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(infos.Peers))
	for _, p := range infos.Peers {
		addr := portRx.ReplaceAllString(p.RemoteAddr, "")
		done := false
		for _, o := range p.Other {
			if strings.HasPrefix(o, "rpc_addr=") == false {
				continue
			}
			done = true
			port := portRx.FindAllString(o, 1)
			if len(port) == 0 {
				return res, fmt.Errorf("Could not extract port from '%s'", o)
			}
			res = append(res, "http://"+addr+port[0])
		}
		if done == false {
			return res, fmt.Errorf("Could not find rpc address")
		}
	}
	return res, nil
}
