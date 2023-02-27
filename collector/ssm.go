package collector

import (
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shatteredsilicon/snmp_exporter/config"
)

type ssmMetric struct {
	Type           string
	RenameTo       string
	HandleLabels   func(labelNames, labelValues []string) ([]string, []string)
	HandleValue    func(value float64) float64
	NewConstMetric func(
		metric *config.Metric,
		t prometheus.ValueType,
		value float64,
		labelNames, labelValues []string,
		constLabels prometheus.Labels,
	) ([]prometheus.Metric, error)
}

var ssmMetrics = map[string]ssmMetric{
	"hrProcessorLoad": {
		Type:         config.MetricTypeGauge,
		RenameTo:     "node_cpu",
		HandleLabels: handleHrDeviceIndex,
	},
	"hrSystemUptime": {
		Type: config.MetricTypeGauge,
		NewConstMetric: func(
			metric *config.Metric,
			t prometheus.ValueType,
			value float64,
			labelNames, labelValues []string,
			constLabels prometheus.Labels,
		) ([]prometheus.Metric, error) {
			seconds := value / 100
			nodeTime := float64(time.Now().Unix())
			nodeBootTime := nodeTime - seconds

			sample1, err := prometheus.NewConstMetric(prometheus.NewDesc("node_time", metric.Help, labelNames, nil),
				t, nodeTime, labelValues...)
			if err != nil {
				return nil, err
			}

			sample2, err := prometheus.NewConstMetric(prometheus.NewDesc("node_boot_time", metric.Help, labelNames, nil),
				t, nodeBootTime, labelValues...)
			if err != nil {
				return nil, err
			}

			return []prometheus.Metric{sample1, sample2}, nil
		},
	},
}

// handleHrDeviceIndex renames label "hrDeviceIndex" to "cpu"
func handleHrDeviceIndex(labelNames, labelValues []string) ([]string, []string) {
	lns := make([]string, len(labelNames))
	for i, name := range labelNames {
		if name == "hrDeviceIndex" {
			lns[i] = "cpu"
		} else {
			lns[i] = name
		}
	}
	return lns, labelValues
}

func isSSMMetrics(metric *config.Metric) bool {
	if metric == nil {
		return false
	}

	sm, ok := ssmMetrics[metric.Name]
	return ok && sm.Type == metric.Type
}

func newSSMConstMetric(
	metric *config.Metric,
	t prometheus.ValueType,
	value float64,
	labelNames, labelValues []string,
	constLabels prometheus.Labels,
) ([]prometheus.Metric, error) {
	if !isSSMMetrics(metric) {
		return nil, errors.New("Not a SSM metric")
	}

	m := ssmMetrics[metric.Name]
	name := metric.Name
	if m.RenameTo != "" {
		name = m.RenameTo
	}
	if m.HandleLabels != nil {
		labelNames, labelValues = m.HandleLabels(labelNames, labelValues)
	}
	if m.HandleValue != nil {
		value = m.HandleValue(value)
	}
	if m.NewConstMetric != nil {
		return m.NewConstMetric(metric, t, value, labelNames, labelValues, constLabels)
	}

	sample, err := prometheus.NewConstMetric(prometheus.NewDesc(name, metric.Help, labelNames, nil),
		t, value, labelValues...)
	return []prometheus.Metric{sample}, err
}
