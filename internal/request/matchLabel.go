package request

import (
	"aurora/internal/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func (wr *WorkerResponse) MatchLabel(labelSelecotr map[string]string) bool {
	if len(labelSelecotr) == 0 {
		return false
	}
	for k, v := range labelSelecotr {
		if vv, ok := wr.Labels[k]; !ok || vv != v {
			return false
		}
	}
	return true
}

// 检测队列是否有订阅者，无则无效
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
	if jsonTable.Messages != 0 {
		return false
	}
	if jsonTable.Consumers == 0 {
		return false
	}
	return true
}
