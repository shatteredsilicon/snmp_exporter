package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	nodeCPUAverageName = "node_cpu_average"
	nodeCPUAverageHelp = "The percentage of CPU utilization."
)

func (c *collector) collecSSMCPUMetrics() ([]prometheus.Metric, error) {
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

	totalTicks := current.totalCPUTicks() - history.totalCPUTicks()
	if totalTicks <= 0 {
		return samples, nil
	}
	labelNames := []string{"cpu", "mode"}
	labelValues := []string{"All"}

	ssCPURawUserDiff := current.ssCPURawUser - history.ssCPURawUser
	if ssCPURawUserDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawUserDiff/totalTicks)*100, append(append([]string{}, labelValues...), "user")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawNiceDiff := current.ssCPURawNice - history.ssCPURawNice
	if ssCPURawNiceDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawNiceDiff/totalTicks)*100, append(append([]string{}, labelValues...), "nice")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawSystemDiff := current.ssCPURawSystem - history.ssCPURawSystem
	if ssCPURawSystemDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawSystemDiff/totalTicks)*100, append(append([]string{}, labelValues...), "system")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawIdleDiff := current.ssCPURawIdle - history.ssCPURawIdle
	if ssCPURawIdleDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawIdleDiff/totalTicks)*100, append(append([]string{}, labelValues...), "idle")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawWaitDiff := current.ssCPURawWait - history.ssCPURawWait
	if ssCPURawWaitDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawWaitDiff/totalTicks)*100, append(append([]string{}, labelValues...), "wait")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawKernelDiff := current.ssCPURawKernel - history.ssCPURawKernel
	if ssCPURawKernelDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawKernelDiff/totalTicks)*100, append(append([]string{}, labelValues...), "kernel")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawInterruptDiff := current.ssCPURawInterrupt - history.ssCPURawInterrupt
	if ssCPURawInterruptDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawInterruptDiff/totalTicks)*100, append(append([]string{}, labelValues...), "irq")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawSoftIRQDiff := current.ssCPURawSoftIRQ - history.ssCPURawSoftIRQ
	if ssCPURawSoftIRQDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawSoftIRQDiff/totalTicks)*100, append(append([]string{}, labelValues...), "softirq")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawStealDiff := current.ssCPURawSteal - history.ssCPURawSteal
	if ssCPURawStealDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawStealDiff/totalTicks)*100, append(append([]string{}, labelValues...), "steal")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}
	ssCPURawGuestDiff := current.ssCPURawGuest - history.ssCPURawGuest
	if ssCPURawGuestDiff >= 0 {
		sample, err := prometheus.NewConstMetric(prometheus.NewDesc(nodeCPUAverageName, nodeCPUAverageHelp, labelNames, nil),
			prometheus.GaugeValue, (ssCPURawGuestDiff/totalTicks)*100, append(append([]string{}, labelValues...), "guest")...)
		if err != nil {
			return samples, err
		}
		samples = append(samples, sample)
	}

	return samples, nil
}