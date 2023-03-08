package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["diskIONRead"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_sectors_read",
		HandleLabels: handlediskIODeviceLabel,
		HandleValue:  handleDiskIOSectorValue,
	}
	ssmMetrics["diskIONWritten"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_sectors_written",
		HandleLabels: handlediskIODeviceLabel,
		HandleValue:  handleDiskIOSectorValue,
	}
	ssmMetrics["diskIONReadX"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_sectors_read",
		HandleLabels: handlediskIODeviceLabel,
		HandleValue:  handleDiskIOSectorValue,
	}
	ssmMetrics["diskIONWrittenX"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_sectors_written",
		HandleLabels: handlediskIODeviceLabel,
		HandleValue:  handleDiskIOSectorValue,
	}
	ssmMetrics["diskIOReads"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_reads_completed",
		HandleLabels: handlediskIODeviceLabel,
	}
	ssmMetrics["diskIOWrites"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_writes_completed",
		HandleLabels: handlediskIODeviceLabel,
	}
	ssmMetrics["diskIOBusyTime"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_disk_io_time_ms",
		HandleLabels: handlediskIODeviceLabel,
		HandleValue:  handleDiskIOBusyTimeValue,
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

func handleDiskIOBusyTimeValue(value float64) float64 { return value / 1000 }

func handleDiskIOSectorValue(value float64) float64 { return value / 512 }
