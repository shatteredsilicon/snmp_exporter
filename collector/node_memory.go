package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shatteredsilicon/snmp_exporter/config"
)

const (
	memAvailableName = "node_memory_MemAvailable"
	memAvailableHelp = "The amount of memory currently available."
	memUsedName      = "node_memory_MemUsed"
	memUsedHelp      = "The amount of memory currently used"
	memVirtualName   = "node_memory_MemVirtual"
	memVirtualHelp   = "The amount of virtual memory"
)

func init() {
	ssmMetrics["memTotalSwap"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_SwapTotal",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["memAvailSwap"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_SwapFree",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["hrMemorySize"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_MemTotal",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["memAvailReal"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_MemFree",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["memShared"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_Shmem",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["memBuffer"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_Buffers",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["memCached"] = ssmMetric{
		Type:        config.MetricTypeGauge,
		RenameTo:    "node_memory_Cached",
		HandleValue: handleMemoryValue,
	}
	ssmMetrics["memSysAvail"] = ssmMetric{
		Type:        config.MetricTypeCounter,
		RenameTo:    "node_memory_MemAvailable",
		HandleValue: handleMemoryValue,
	}
}

// handleMemoryValue converts memory unit from 'KB' to 'B'
func handleMemoryValue(value float64) float64 { return value * 1024 }

func (c *Collector) collectSSMMemoryMetrics() ([]prometheus.Metric, error) {
	samples := []prometheus.Metric{}

	current, ok := ssmMetricRecords.current[c.target]
	if !ok {
		return samples, nil
	}

	if _, ok := current.collectedMetrics["memAvailReal"]; ok {
		return samples, nil
	}

	sample, err := prometheus.NewConstMetric(prometheus.NewDesc(memUsedName, memUsedHelp, nil, nil),
		prometheus.GaugeValue, current.hrSWRunPerfMem*1024)
	if err != nil {
		return samples, err
	}
	samples = append(samples, sample)

	memAvail := (current.hrMemorySize - current.hrSWRunPerfMem) * 1024 // convert KB to B
	sample, err = prometheus.NewConstMetric(prometheus.NewDesc(memAvailableName, memAvailableHelp, nil, nil),
		prometheus.GaugeValue, memAvail)
	if err != nil {
		return samples, err
	}
	samples = append(samples, sample)

	return samples, nil
}
