package metrics

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// 公共接口，提供给三方调用
// 自动添加累加指标
func IncCounterWithLabelValues(index int, lvs ...string) error {
	counterVec, err := getCounterVec(index)
	if err != nil {
		return err
	}
	vals := getVals(lvs...)
	counterVec.WithLabelValues(vals...).Inc()
	return nil
}

// 添加累加指标
func AddCounterWithLabelValues(number float64, index int, lvs ...string) error {
	counterVec, err := getCounterVec(index)
	if err != nil {
		return err
	}
	vals := getVals(lvs...)
	counterVec.WithLabelValues(vals...).Add(number)
	return nil
}

// 添加测量指标
func AddGaugeWithLabelValues(number float64, index int, lvs ...string) error {
	gaugeVec, err := getGaugeVec(index)
	if err != nil {
		return err
	}
	vals := getVals(lvs...)
	gaugeVec.WithLabelValues(vals...).Add(number)
	return nil
}

// 设置测量指标
func SetGaugeWithLabelValues(number float64, index int, lvs ...string) error {
	gaugeVec, err := getGaugeVec(index)
	if err != nil {
		return err
	}
	vals := getVals(lvs...)
	gaugeVec.WithLabelValues(vals...).Set(number)
	return nil
}

// 添加直方图指标
func AddHistogramWithLabelValues(number float64, index int, lvs ...string) error {
	histogramVec, err := getHistogramVec(index)
	if err != nil {
		return err
	}
	vals := getVals(lvs...)
	histogramVec.WithLabelValues(vals...).Observe(number)
	return nil
}

// 添加概略图指标
func AddSummaryWithLabelValues(number float64, index int, lvs ...string) error {
	summaryVec, err := getSummaryVec(index)
	if err != nil {
		return err
	}
	vals := getVals(lvs...)
	summaryVec.WithLabelValues(vals...).Observe(number)
	return nil
}

func getCounterVec(index int) (*prometheus.CounterVec, error) {
	mc := CollectorrRegistry[index]
	counterVec, ok := mc.(*prometheus.CounterVec)
	if !ok {
		return nil, errors.New("counter with other metric type conflicts")
	}
	return counterVec, nil
}

func getGaugeVec(index int) (*prometheus.GaugeVec, error) {
	mc := CollectorrRegistry[index]
	gaugeVec, ok := mc.(*prometheus.GaugeVec)
	if !ok {
		return nil, errors.New("gauge with other metric type conflicts")
	}
	return gaugeVec, nil
}

func getHistogramVec(index int) (*prometheus.HistogramVec, error) {
	mc := CollectorrRegistry[index]
	histogramVec, ok := mc.(*prometheus.HistogramVec)
	if !ok {
		return nil, errors.New("histogram with other metric type conflicts")
	}
	return histogramVec, nil
}

func getSummaryVec(index int) (*prometheus.SummaryVec, error) {
	mc := CollectorrRegistry[index]
	summaryVec, ok := mc.(*prometheus.SummaryVec)
	if !ok {
		return nil, errors.New("summary with other metric type conflicts")
	}
	return summaryVec, nil
}

func getVals(lvs ...string) []string {
	vals := []string{Region, Hostname, AppVersion}
	vals = append(vals, lvs...)
	return vals
}
