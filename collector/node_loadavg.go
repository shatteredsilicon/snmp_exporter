package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	nodeLoad1Name  = "node_load1"
	nodeLoad5Name  = "node_load5"
	nodeLoad15Name = "node_load10"

	nodeLoad1Help  = "1m load average."
	nodeLoad5Help  = "5m load average."
	nodeLoad15Help = "15m load average."

	nodeLoad5Duration  float64 = 4 * 60.0
	nodeLoad15Duration float64 = 14 * 60.0
)

func (c *Collector) collectSSMLoadavgMetrics() ([]prometheus.Metric, error) {
	samples := []prometheus.Metric{}

	history, ok := ssmMetricRecords.history[c.target]
	if !ok {
		return samples, nil
	}
	current, ok := ssmMetricRecords.current[c.target]
	if !ok {
		return samples, nil
	}
	if current.hrSystemDate <= history.hrSystemDate {
		return samples, nil
	}

	loadavg := c.currentProcessorLoad()
	if loadavg == -1 {
		return samples, nil
	}

	// calculate loadavg1, loadavg5, loadavg10 value
	var load5Total, load15Total float64 = loadavg, loadavg
	var load5Count, load15Count int = 1, 1
	for i := len(history.hrProcessorLoads) - 1; i >= 0; i-- {
		load := history.hrProcessorLoads[i]

		if (current.hrSystemDate - load.hrSystemDate) < nodeLoad5Duration {
			load5Total += load.value
			load5Count += 1
		}
		if (current.hrSystemDate - load.hrSystemDate) >= nodeLoad15Duration {
			break
		}
		load15Total += load.value
		load15Count += 1
	}

	sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeLoad1Name, nodeLoad1Help, nil, nil),
		prometheus.GaugeValue, loadavg)
	if err != nil {
		return samples, err
	}
	samples = append(samples, sample)

	sample, err = prometheus.NewConstMetric(prometheus.NewDesc(nodeLoad5Name, nodeLoad5Help, nil, nil),
		prometheus.GaugeValue, load5Total/float64(load5Count))
	if err != nil {
		return samples, err
	}
	samples = append(samples, sample)

	sample, err = prometheus.NewConstMetric(prometheus.NewDesc(nodeLoad15Name, nodeLoad5Help, nil, nil),
		prometheus.GaugeValue, load15Total/float64(load15Count))
	if err != nil {
		return samples, err
	}
	samples = append(samples, sample)

	return samples, nil
}
