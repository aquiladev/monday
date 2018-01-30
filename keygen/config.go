package keygen

import (
	"encoding/json"
	"io/ioutil"

	"github.com/aquiladev/monday/util"
)

type Range struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type RangeConfig struct {
	Range Range `json:"range"`
}

func fetchRange(url string) (*RangeConfig, error) {
	resp, err := util.Get(url)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var config RangeConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
