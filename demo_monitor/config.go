package main

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	Nodes []node
}

func loadConfig(path string) (config, error) {
	res := config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(data, &res)
}
