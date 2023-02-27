package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["ssCpuRawUser"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("user"),
	}
	ssmMetrics["ssCpuRawNice"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("nice"),
	}
	ssmMetrics["ssCpuRawSystem"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("system"),
	}
	ssmMetrics["ssCpuRawIdle"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("idle"),
	}
	ssmMetrics["ssCpuRawWait"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("iowait"),
	}
	ssmMetrics["ssCpuRawInterrupt"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("irq"),
	}
	ssmMetrics["ssCpuRawSoftIRQ"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("softirq"),
	}
	ssmMetrics["ssCpuRawSteal"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("steal"),
	}
	ssmMetrics["ssCpuRawGuest"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_cpu",
		HandleLabels: cpuStatsHandler("guest"),
	}
}

func cpuStatsHandler(mode string) func([]string, []string) ([]string, []string) {
	return func(labelNames, labelValues []string) ([]string, []string) {
		lns := append(append([]string{}, labelNames...), "cpu", "mode")
		lvs := append(append([]string{}, labelValues...), "cpu", mode)
		return lns, lvs
	}
}
