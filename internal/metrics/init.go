package metrics

import (
	"errors"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Region             string
	AppVersion         string
	Ip                 string
	Hostname           string
	CollectorrRegistry []prometheus.Collector
)

func init() {
	Hostname, _ = os.Hostname()
}

func InitMetrics(region, version, exporterPath, watcherPath string) error {
	Region = region
	AppVersion = "0.0.0"
	if version != "" {
		AppVersion = version
	}
	exporterConfig, err := ParseExporterConfig(exporterPath)
	if err != nil {
		return err
	}
	initDataItems(exporterConfig.Metrics)
	return initExporterMetrics(Region, AppVersion, Hostname, exporterConfig)
}

func initExporterMetrics(region, version, hostname string, exporterConfig ExporterConfig) error {
	if err := regiterMetrics(exporterConfig.Metrics); err != nil {
		return err
	}
	exporter := NewSysExporter(region, version, hostname, exporterConfig.Metrics)
	prometheus.MustRegister(exporter)
	return nil
}

func initDataItems(metricsConfig MetricsConfig) {
	maxIndex := getMaxIndex(metricsConfig.Datas)
	MetricsDatas = make([]DataItem, maxIndex+1)
	for _, v := range metricsConfig.Datas {
		MetricsDatas[v.Index] = v
	}
}

func regiterMetrics(metricsConfig MetricsConfig) error {
	maxIndex := getMaxIndex(metricsConfig.Datas)
	CollectorrRegistry = make([]prometheus.Collector, maxIndex+1)
	for _, v := range metricsConfig.Datas {
		labels := []string{"region", "hostname", "version"}
		if len(v.Labels) != 0 {
			lbs := strings.Split(v.Labels, "-")
			labels = append(labels, lbs...)
		}

		switch v.Type {
		case METRICS_TYPE_COUNTER:
			promCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: metricsConfig.Namespace,
				Subsystem: metricsConfig.Subsystem,
				Name:      v.Name,
				Help:      v.Help,
			}, labels)

			prometheus.MustRegister(promCounter)
			CollectorrRegistry[v.Index] = promCounter
		case METRICS_TYPE_GAUGE:
			promGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: metricsConfig.Namespace,
				Subsystem: metricsConfig.Subsystem,
				Name:      v.Name,
				Help:      v.Help,
			}, labels)

			prometheus.MustRegister(promGauge)
			CollectorrRegistry[v.Index] = promGauge
		case METRICS_TYPE_HISTOGRAM:
			promHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: metricsConfig.Namespace,
				Subsystem: metricsConfig.Subsystem,
				Name:      v.Name,
				Help:      v.Help,
				Buckets:   v.Buckets,
			}, labels)

			prometheus.MustRegister(promHistogram)
			CollectorrRegistry[v.Index] = promHistogram
		case METRICS_TYPE_SUMMARY:
			promSummary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Namespace:  metricsConfig.Namespace,
				Subsystem:  metricsConfig.Subsystem,
				Name:       v.Name,
				Help:       v.Help,
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 响应时间的分位
			}, labels)

			prometheus.MustRegister(promSummary)
			CollectorrRegistry[v.Index] = promSummary
		default:
			return errors.New("metrics type is invalid")
		}
	}
	return nil
}

func getMaxIndex(datas []DataItem) int {
	maxIndex := 0
	for _, v := range datas {
		if v.Index > maxIndex {
			maxIndex = v.Index
		}
	}
	return maxIndex
}

func clear() {
	if len(CollectorrRegistry) != 0 {
		for _, cr := range CollectorrRegistry {
			prometheus.Unregister(cr)
		}
		CollectorrRegistry = CollectorrRegistry[:0]
	}
}

func Stop() {
	clear()
}
