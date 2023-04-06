package metrics

import "github.com/spf13/viper"

type MetricsConfig struct {
	Namespace string
	Subsystem string
	Datas     []DataItem
}

type ExporterConfig struct {
	Metrics MetricsConfig
}

type DataItem struct {
	Index   int
	Type    int
	Name    string
	Help    string
	Buckets []float64
	Labels  string
}

var (
	DefaultExporterFilePath = "./config/metrics.toml"
	MetricsDatas            []DataItem
	IsInitMetrics           bool
)

func ParseExporterConfig(exporterPath string) (ExporterConfig, error) {
	if len(exporterPath) == 0 {
		exporterPath = DefaultExporterFilePath
	}
	var exporterConfig ExporterConfig
	if err := decodeConfig(exporterPath, &exporterConfig); err != nil {
		return exporterConfig, err
	}
	return exporterConfig, nil
}

func decodeConfig(filePath string, conf interface{}) (err error) {
	configViper := viper.New()
	configViper.SetConfigFile(filePath)
	if err := configViper.ReadInConfig(); err != nil {
		return err
	}
	if err = configViper.Unmarshal(conf); err != nil {
		return err
	}
	return nil
}

func GetIndexDataKey(index int) string {
	for _, v := range MetricsDatas {
		if index == v.Index {
			return v.Name + "-" + v.Labels
		}
	}
	return ""
}
