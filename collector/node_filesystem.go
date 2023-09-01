package collector

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shatteredsilicon/snmp_exporter/config"
)

const (
	hrStorageVirtualMemory = "3"
	hrStorageFixedDisk     = "4"

	filesystemSizeName = "node_filesystem_size"
	filesystemUsedName = "node_filesystem_used"
	filesystemFreeName = "node_filesystem_free"

	filesystemUsedHelp = "The used size of the filesystem"
	filesystemFreeHelp = "The free size of the filesystem"
)

func init() {
	ssmMetrics["hrStorageSize"] = ssmMetric{
		Type: config.MetricTypeGauge,
		NewConstMetric: func(
			metric *config.Metric,
			t prometheus.ValueType,
			value float64,
			labelNames, labelValues []string,
			constLabels prometheus.Labels,
		) ([]prometheus.Metric, error) {
			var unit, used float64
			var typ, descr string
			var err error

			metricLNs := []string{"device", "fstype", "mountpoint"}
			metricLVs := []string{"", "unknown"}

			for i, labelName := range labelNames {
				switch labelName {
				case "hrStorageAllocationUnits":
					unit, err = strconv.ParseFloat(labelValues[i], 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse hrStorageAllocationUnits: %s", err.Error())
					}
				case "hrStorageType":
					parts := strings.Split(labelValues[i], ".")
					typ = parts[len(parts)-1]
				case "hrStorageDescr":
					descr = labelValues[i]
					labelIndex := strings.Index(descr, " Label:")
					if labelIndex != -1 {
						descr = descr[:labelIndex]
					}
					metricLVs = append(metricLVs, descr)
				case "hrStorageUsed":
					used, err = strconv.ParseFloat(labelValues[i], 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse hrStorageUsed: %s", err.Error())
					}
				}
			}

			samples := []prometheus.Metric{}
			if typ == hrStorageVirtualMemory && strings.ToLower(strings.TrimSpace(descr)) == "virtual memory" {
				sample, err := prometheus.NewConstMetric(prometheus.NewDesc(memVirtualName, memVirtualHelp, nil, nil),
					t, unit*value)
				if err != nil {
					return samples, err
				}
				samples = append(samples, sample)
			}

			if typ == hrStorageFixedDisk {
				sample, err := prometheus.NewConstMetric(prometheus.NewDesc(filesystemSizeName, removeOidSuffix(metric.Help), metricLNs, nil),
					t, value*unit, metricLVs...)
				if err != nil {
					return samples, err
				}
				samples = append(samples, sample)

				sample, err = prometheus.NewConstMetric(prometheus.NewDesc(filesystemUsedName, filesystemUsedHelp, metricLNs, nil),
					t, used*unit, metricLVs...)
				if err != nil {
					return samples, err
				}
				samples = append(samples, sample)

				sample, err = prometheus.NewConstMetric(prometheus.NewDesc(filesystemFreeName, filesystemFreeHelp, metricLNs, nil),
					t, (value-used)*unit, metricLVs...)
				if err != nil {
					return samples, err
				}
				samples = append(samples, sample)
			}

			return samples, nil
		},
	}
}
