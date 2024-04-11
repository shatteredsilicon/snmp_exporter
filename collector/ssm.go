package collector

import (
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shatteredsilicon/snmp_exporter/config"
)

type ssmMetricProcessorLoad struct {
	hrSystemDate float64
	value        float64
}

type ssmMetricPerfCPU struct {
	hrSWRunType string
	value       float64
}

type ssmMetricRecord struct {
	ssCPURawUser      float64
	ssCPURawNice      float64
	ssCPURawSystem    float64
	ssCPURawIdle      float64
	ssCPURawWait      float64
	ssCPURawKernel    float64
	ssCPURawInterrupt float64
	ssCPURawSoftIRQ   float64
	ssCPURawSteal     float64
	ssCPURawGuest     float64
	hrSystemDate      float64
	hrProcessorLoad   []float64
	hrProcessorLoads  []ssmMetricProcessorLoad
	hrMemorySize      float64
	hrSWRunPerfMem    float64
	hrSWRunPerfCPU    map[string]ssmMetricPerfCPU
	hrSWRunName       map[string]string
	collectedMetrics  map[string]struct{}
	mu                sync.Mutex
}

func (r *ssmMetricRecord) totalCPUTicks() float64 {
	return r.ssCPURawUser +
		r.ssCPURawNice +
		r.ssCPURawSystem +
		r.ssCPURawIdle +
		r.ssCPURawWait +
		r.ssCPURawKernel +
		r.ssCPURawInterrupt +
		r.ssCPURawSoftIRQ +
		r.ssCPURawSteal +
		r.ssCPURawGuest
}

var ssmMetricRecords = struct {
	history map[string]*ssmMetricRecord
	current map[string]*ssmMetricRecord
}{
	history: make(map[string]*ssmMetricRecord),
	current: make(map[string]*ssmMetricRecord),
}

type ssmMetric struct {
	Type           string
	RenameTo       string
	HandleLabels   func(labelNames, labelValues []string) ([]string, []string)
	HandleValue    func(value float64) float64
	Help           string
	NewConstMetric func(
		metric *config.Metric,
		t prometheus.ValueType,
		value float64,
		labelNames, labelValues []string,
		constLabels prometheus.Labels,
	) ([]prometheus.Metric, error)
}

var ssmMetrics = map[string]ssmMetric{
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

			sample1, err := prometheus.NewConstMetric(prometheus.NewDesc("node_time", removeOidSuffix(metric.Help), labelNames, nil),
				t, nodeTime, labelValues...)
			if err != nil {
				return nil, err
			}

			sample2, err := prometheus.NewConstMetric(prometheus.NewDesc("node_boot_time", removeOidSuffix(metric.Help), labelNames, nil),
				t, nodeBootTime, labelValues...)
			if err != nil {
				return nil, err
			}

			return []prometheus.Metric{sample1, sample2}, nil
		},
	},
}

func isSSMMetrics(metric *config.Metric) bool {
	if metric == nil {
		return false
	}

	sm, ok := ssmMetrics[metric.Name]
	return ok && sm.Type == metric.Type
}

func removeOidSuffix(msg string) string {
	oidSuffixR, _ := regexp.Compile(`- (\d+\.)*\d+$`)
	matchedStr := oidSuffixR.FindString(msg)
	return strings.TrimSuffix(msg, matchedStr)
}

func newSSMConstMetric(
	metric *config.Metric,
	t prometheus.ValueType,
	value float64,
	labelNames, labelValues []string,
	constLabels prometheus.Labels,
) ([]prometheus.Metric, error) {
	if !isSSMMetrics(metric) {
		return nil, errors.New("not a SSM metric")
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

	help := removeOidSuffix(metric.Help)
	if m.Help != "" {
		help = m.Help
	}
	sample, err := prometheus.NewConstMetric(prometheus.NewDesc(name, help, labelNames, nil),
		t, value, labelValues...)
	return []prometheus.Metric{sample}, err
}

func (c *Collector) copyHistorySSMMetrics() {
	current, ok := ssmMetricRecords.current[c.target]
	if !ok {
		return
	}

	hrProcessorLoads := []ssmMetricProcessorLoad{}
	history, ok := ssmMetricRecords.history[c.target]
	if ok {
		hrProcessorLoads = history.hrProcessorLoads
	}

	for i := len(hrProcessorLoads) - 1; i >= 0; i-- {
		if (current.hrSystemDate - hrProcessorLoads[i].hrSystemDate) < nodeLoad15Duration {
			continue
		}
		if i == len(hrProcessorLoads)-1 {
			hrProcessorLoads = []ssmMetricProcessorLoad{}
		} else {
			hrProcessorLoads = hrProcessorLoads[i+1:]
		}
		break
	}

	if loadavg := c.currentProcessorLoad(); loadavg != -1 {
		hrProcessorLoads = append(hrProcessorLoads, ssmMetricProcessorLoad{
			hrSystemDate: current.hrSystemDate,
			value:        loadavg,
		})
	}

	ssmMetricRecords.history[c.target] = &ssmMetricRecord{
		ssCPURawUser:      current.ssCPURawUser,
		ssCPURawNice:      current.ssCPURawNice,
		ssCPURawSystem:    current.ssCPURawSystem,
		ssCPURawIdle:      current.ssCPURawIdle,
		ssCPURawWait:      current.ssCPURawWait,
		ssCPURawKernel:    current.ssCPURawKernel,
		ssCPURawInterrupt: current.ssCPURawInterrupt,
		ssCPURawSoftIRQ:   current.ssCPURawSoftIRQ,
		ssCPURawSteal:     current.ssCPURawSteal,
		ssCPURawGuest:     current.ssCPURawGuest,
		hrSystemDate:      current.hrSystemDate,
		hrProcessorLoads:  hrProcessorLoads,
		hrSWRunPerfCPU:    current.hrSWRunPerfCPU,
	}
}

func (c *Collector) currentProcessorLoad() float64 {
	current, ok := ssmMetricRecords.current[c.target]
	if !ok || len(current.hrProcessorLoad) == 0 {
		return -1
	}

	var total float64 = 0
	for _, load := range current.hrProcessorLoad {
		total += load
	}

	return total / float64(len(current.hrProcessorLoad))
}
