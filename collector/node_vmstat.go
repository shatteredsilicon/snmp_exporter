package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["ssIORawReceived"] = ssmMetric{
		Type:        config.MetricTypeCounter,
		RenameTo:    "node_vmstat_pgpgin",
		HandleValue: handleVMStatPageValue,
	}
	ssmMetrics["ssIORawSent"] = ssmMetric{
		Type:        config.MetricTypeCounter,
		RenameTo:    "node_vmstat_pgpgout",
		HandleValue: handleVMStatPageValue,
	}
	ssmMetrics["ssSwapIn"] = ssmMetric{
		Type:     config.MetricTypeGauge,
		RenameTo: "node_vmstat_pswpin",
	}
	ssmMetrics["ssSwapOut"] = ssmMetric{
		Type:     config.MetricTypeGauge,
		RenameTo: "node_vmstat_pswpout",
	}
}

func handleVMStatPageValue(value float64) float64 { return value / 2 }
