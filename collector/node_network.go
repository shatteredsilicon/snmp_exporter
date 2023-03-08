package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["ifInOctets"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_receive_bytes",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifInUcastPkts"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_receive_packets",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifInNUcastPkts"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_receive_multicast",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifInDiscards"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_receive_drop",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifInErrors"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_receive_errs",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifOutOctets"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_transmit_bytes",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifOutUcastPkts"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_transmit_packets",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifOutNUcastPkts"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_transmit_multicast",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifOutDiscards"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_transmit_drop",
		HandleLabels: handleIfDescrLabel,
	}
	ssmMetrics["ifOutErrors"] = ssmMetric{
		Type:         config.MetricTypeCounter,
		RenameTo:     "node_network_transmit_errs",
		HandleLabels: handleIfDescrLabel,
	}
}

// handleIfDescrLabel renames label "ifDescr" to "device"
func handleIfDescrLabel(labelNames, labelValues []string) ([]string, []string) {
	lns := make([]string, len(labelNames))
	for i, name := range labelNames {
		if name == "ifDescr" {
			lns[i] = "device"
		} else {
			lns[i] = name
		}
	}
	return lns, labelValues
}
