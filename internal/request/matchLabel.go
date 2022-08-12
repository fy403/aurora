package request

import (
	"aurora/internal/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func (wr *WorkerResponse) MatchLabel(labelSelecotr map[string]string) bool {
	for k, v := range labelSelecotr {
		if vv, ok := wr.Labels[k]; !ok || vv != v {
			return false
		}
	}
	return true
}

func (wr *WorkerResponse) IsValid(brokerApi string) bool {
	url := brokerApi + wr.SpecQueue
	resp, err := http.Get(url)
	if resp == nil || err != nil {
		log.Runtime().Errorf("isValid func call fail: %s", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Runtime().Errorf("isValid func call fail, StatusCode is %d", resp.StatusCode)
		return false
	}
	jsonTable := RabbitMQApi{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Runtime().Errorf("isValid func call fail: %s", err)
		return false
	}
	err = json.Unmarshal(body, &jsonTable)
	if err != nil {
		log.Runtime().Errorf("isValid func call fail: %s", err)
		return false
	}
	if jsonTable.Consumers == 0 {
		return false
	}
	return true
}
