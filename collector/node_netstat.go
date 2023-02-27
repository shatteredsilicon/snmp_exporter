package collector

import "github.com/shatteredsilicon/snmp_exporter/config"

func init() {
	ssmMetrics["ipSystemStatsInReceives"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InReceives",
	}
	ssmMetrics["ipSystemStatsHCInReceives"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InReceives",
	}
	ssmMetrics["ipSystemStatsInOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InOctets",
	}
	ssmMetrics["ipSystemStatsHCInOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InOctets",
	}
	ssmMetrics["ipSystemStatsInHdrErrors"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IP_InHdrErrors",
	}
	ssmMetrics["ipSystemStatsInNoRoutes"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InNoRoutes",
	}
	ssmMetrics["ipSystemStatsInAddrErrors"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InAddrErrors",
	}
	ssmMetrics["ipSystemStatsInUnknownProtos"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InUnknownProtos",
	}
	ssmMetrics["ipSystemStatsInTruncatedPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InTruncatedPkts",
	}
	ssmMetrics["ipSystemStatsReasmReqds"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_ReasmReqds",
	}
	ssmMetrics["ipSystemStatsReasmOKs"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_ReasmOKs",
	}
	ssmMetrics["ipSystemStatsReasmFails"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_ReasmFails",
	}
	ssmMetrics["ipSystemStatsInDiscards"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InDiscards",
	}
	ssmMetrics["ipSystemStatsInDelivers"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InDelivers",
	}
	ssmMetrics["ipSystemStatsHCInDelivers"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_InDelivers",
	}
	ssmMetrics["ipSystemStatsOutRequests"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_OutRequests",
	}
	ssmMetrics["ipSystemStatsHCOutRequests"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_OutRequests",
	}
	ssmMetrics["ipSystemStatsOutNoRoutes"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_OutNoRoutes",
	}
	ssmMetrics["ipSystemStatsOutForwDatagrams"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_ForwDatagrams",
	}
	ssmMetrics["ipSystemStatsHCOutForwDatagrams"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_ForwDatagrams",
	}
	ssmMetrics["ipSystemStatsOutDiscards"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_OutDiscards",
	}
	ssmMetrics["ipSystemStatsOutFragOKs"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_FragOKs",
	}
	ssmMetrics["ipSystemStatsOutFragFails"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_FragFails",
	}
	ssmMetrics["ipSystemStatsOutFragCreates"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_Ip_FragCreates",
	}
	ssmMetrics["ipSystemStatsOutOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutOctets",
	}
	ssmMetrics["ipSystemStatsHCOutOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutOctets",
	}
	ssmMetrics["ipSystemStatsInMcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InMcastPkts",
	}
	ssmMetrics["ipSystemStatsHCInMcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InMcastPkts",
	}
	ssmMetrics["ipSystemStatsInMcastOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InMcastOctets",
	}
	ssmMetrics["ipSystemStatsHCInMcastOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InMcastOctets",
	}
	ssmMetrics["ipSystemStatsOutMcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutMcastPkts",
	}
	ssmMetrics["ipSystemStatsHCOutMcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutMcastPkts",
	}
	ssmMetrics["ipSystemStatsOutMcastOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutMcastOctets",
	}
	ssmMetrics["ipSystemStatsHCOutMcastOctets"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutMcastOctets",
	}
	ssmMetrics["ipSystemStatsInBcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InBcastPkts",
	}
	ssmMetrics["ipSystemStatsHCInBcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_InBcastPkts",
	}
	ssmMetrics["ipSystemStatsOutBcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutBcastPkts",
	}
	ssmMetrics["ipSystemStatsHCOutBcastPkts"] = ssmMetric{
		Type:     config.MetricTypeCounter,
		RenameTo: "node_netstat_IpExt_OutBcastPkts",
	}
}
