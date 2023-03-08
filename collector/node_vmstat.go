package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["ssIORawReceived"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_vmstat_pgpgin",
	}
	ssmMetrics["ssIORawSent"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_vmstat_pgpgout",
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
