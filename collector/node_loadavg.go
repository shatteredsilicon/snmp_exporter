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
			lns := make([]string, 0)
			lvs := make([]string, 0)
			for i := range labelNames {
				if labelNames[i] != "laIndex" {
					lns = append(lns, labelNames[i])
					lvs = append(lvs, labelValues[i])
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
			sample, err := prometheus.NewConstMetric(prometheus.NewDesc(name, removeOidSuffix(metric.Help), lns, nil),
				t, value, lvs...)
			if err != nil {
				return nil, err
			}

			return []prometheus.Metric{sample}, nil
		},
	}
}
