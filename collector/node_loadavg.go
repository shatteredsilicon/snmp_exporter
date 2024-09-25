package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shatteredsilicon/snmp_exporter/config"
)

const (
	nodeLoad1Name  = "node_load1"
	nodeLoad5Name  = "node_load5"
	nodeLoad15Name = "node_load15"

	nodeLoad1Help  = "1m load average."
	nodeLoad5Help  = "5m load average."
	nodeLoad15Help = "15m load average."
)

func init() {
	ssmMetrics["laLoadFloat"] = ssmMetric{
		Type: config.MetricTypeFloat,
		NewConstMetric: func(
			metric *config.Metric,
			t prometheus.ValueType,
			value float64,
			labelNames, labelValues []string,
			constLabels prometheus.Labels,
		) ([]prometheus.Metric, error) {
			for i, labelName := range labelNames {
				if labelName != "laNames" || i >= len(labelValues) {
					break
				}

				switch labelValues[i] {
				case "Load-1":
					sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeLoad1Name, nodeLoad1Help, nil, nil),
						t, value)
					if err != nil {
						return nil, err
					}
					return []prometheus.Metric{sample}, nil
				case "Load-5":
					sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeLoad5Name, nodeLoad5Help, nil, nil),
						t, value)
					if err != nil {
						return nil, err
					}
					return []prometheus.Metric{sample}, nil
				case "Load-15":
					sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeLoad15Name, nodeLoad15Help, nil, nil),
						t, value)
					if err != nil {
						return nil, err
					}
					return []prometheus.Metric{sample}, nil
				}
			}

			return []prometheus.Metric{}, nil
		},
	}
}
