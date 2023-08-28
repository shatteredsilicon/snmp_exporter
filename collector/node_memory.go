package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

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
