package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["diskIONRead"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_disk_sectors_read",
	}
	ssmMetrics["diskIONWritten"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_disk_sectors_written",
	}
	ssmMetrics["diskIONReadX"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_disk_sectors_read",
	}
	ssmMetrics["diskIONWrittenX"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_disk_sectors_written",
	}
	ssmMetrics["diskIOReads"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_disk_reads_completed",
	}
	ssmMetrics["diskIOWrites"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_disk_writes_completed",
	}
}

// handlediskIODeviceLabel renames label "ifDescr" to "device"
func handlediskIODeviceLabel(labelNames, labelValues []string) ([]string, []string) {
	lns := make([]string, len(labelNames))
	for i, name := range labelNames {
		if name == "diskIODevice" {
			lns[i] = "device"
		} else {
			lns[i] = name
		}
	}
	return lns, labelValues
}
