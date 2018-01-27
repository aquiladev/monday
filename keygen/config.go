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

func mockConfig() *RangeConfig {
	return &RangeConfig{
		Range: Range{
			From: "100152338825365595862742132647329357860924845607696427128849574371092698716",
			To:   "100152338825365595862742132647329357860924845607696427128849574371092698815",
		},
	}
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
