package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shatteredsilicon/snmp_exporter/config"
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
			name := "node_load"
			for i := range labelNames {
				if labelNames[i] != "laIndex" {
					continue
				}
				if len(labelValues) <= i {
					break
				}
				if labelValues[i] == "1" {
					name += "1"
				} else if labelValues[i] == "2" {
					name += "5"
				} else {
					name += "15"
				}
			}
			sample, err := prometheus.NewConstMetric(prometheus.NewDesc(name, removeOidSuffix(metric.Help), labelNames, nil),
				t, value, labelValues...)
			if err != nil {
				return nil, err
			}

			return []prometheus.Metric{sample}, nil
		},
	}
}
