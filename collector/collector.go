// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gosnmp/gosnmp"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/shatteredsilicon/snmp_exporter/config"
)

var (
	// 64-bit float mantissa: https://en.wikipedia.org/wiki/Double-precision_floating-point_format
	float64Mantissa uint64 = 9007199254740992
	wrapCounters           = kingpin.Flag("snmp.wrap-large-counters", "Wrap 64-bit counters to avoid floating point rounding.").Default("true").Bool()
	srcAddress             = kingpin.Flag("snmp.source-address", "Source address to send snmp from in the format 'address:port' to use when connecting targets. If the port parameter is empty or '0', as in '127.0.0.1:' or '[::1]:0', a source port number is automatically (random) chosen.").Default("").String()
)

// Types preceded by an enum with their actual type.
var combinedTypeMapping = map[string]map[int]string{
	"InetAddress": {
		1: "InetAddressIPv4",
		2: "InetAddressIPv6",
	},
	"InetAddressMissingSize": {
		1: "InetAddressIPv4",
		2: "InetAddressIPv6",
	},
	"LldpPortId": {
		1: "DisplayString",
		2: "DisplayString",
		3: "PhysAddress48",
		5: "DisplayString",
		7: "DisplayString",
	},
}

func oidToList(oid string) []int {
	result := []int{}
	for _, x := range strings.Split(oid, ".") {
		o, _ := strconv.Atoi(x)
		result = append(result, o)
	}
	return result
}

func listToOid(l []int) string {
	var result []string
	for _, o := range l {
		result = append(result, strconv.Itoa(o))
	}
	return strings.Join(result, ".")
}

type ScrapeResults struct {
	pdus    []gosnmp.SnmpPDU
	packets uint64
	retries uint64
}

func ScrapeTarget(ctx context.Context, target string, auth *config.Auth, module *config.Module, logger log.Logger, metrics Metrics) (ScrapeResults, error) {
	results := ScrapeResults{}
	// Set the options.
	snmp := gosnmp.GoSNMP{}
	snmp.Context = ctx
	snmp.MaxRepetitions = module.WalkParams.MaxRepetitions
	snmp.Retries = *module.WalkParams.Retries
	snmp.Timeout = module.WalkParams.Timeout
	snmp.UseUnconnectedUDPSocket = module.WalkParams.UseUnconnectedUDPSocket
	snmp.LocalAddr = *srcAddress

	// Allow a set of OIDs that aren't in a strictly increasing order
	if module.WalkParams.AllowNonIncreasingOIDs {
		snmp.AppOpts = make(map[string]interface{})
		snmp.AppOpts["c"] = true
	}

	var sent time.Time
	snmp.OnSent = func(x *gosnmp.GoSNMP) {
		sent = time.Now()
		metrics.SNMPPackets.Inc()
		results.packets++
	}
	snmp.OnRecv = func(x *gosnmp.GoSNMP) {
		metrics.SNMPDuration.Observe(time.Since(sent).Seconds())
	}
	snmp.OnRetry = func(x *gosnmp.GoSNMP) {
		metrics.SNMPRetries.Inc()
		results.retries++
	}

	// Configure target.
	if err := configureTarget(&snmp, target); err != nil {
		return results, err
	}

	// Configure auth.
	auth.ConfigureSNMP(&snmp)

	// Do the actual walk.
	getInitialStart := time.Now()
	err := snmp.Connect()
	if err != nil {
		if err == context.Canceled {
			return results, fmt.Errorf("scrape cancelled after %s (possible timeout) connecting to target %s",
				time.Since(getInitialStart), snmp.Target)
		}
		return results, fmt.Errorf("error connecting to target %s: %s", target, err)
	}
	defer snmp.Conn.Close()

	// Evaluate rules.
	newGet := module.Get
	newWalk := module.Walk
	for _, filter := range module.Filters {
		var pdus []gosnmp.SnmpPDU
		allowedList := []string{}

		if snmp.Version == gosnmp.Version1 {
			pdus, err = snmp.WalkAll(filter.Oid)
		} else {
			pdus, err = snmp.BulkWalkAll(filter.Oid)
		}
		// Do not try to filter anything if we had errors.
		if err != nil {
			level.Info(logger).Log("msg", "Error getting OID, won't do any filter on this oid", "oid", filter.Oid)
			continue
		}

		allowedList = filterAllowedIndices(logger, filter, pdus, allowedList, metrics)

		// Update config to get only index and not walk them.
		newWalk = updateWalkConfig(newWalk, filter, logger)

		// Only Keep indices not involved in filters.
		newCfg := updateGetConfig(newGet, filter, logger)

		// We now add each index from filter to the get list.
		newCfg = addAllowedIndices(filter, allowedList, logger, newCfg)

		newGet = newCfg
	}

	getOids := newGet
	maxOids := int(module.WalkParams.MaxRepetitions)
	// Max Repetition can be 0, maxOids cannot. SNMPv1 can only report one OID error per call.
	if maxOids == 0 || snmp.Version == gosnmp.Version1 {
		maxOids = 1
	}
	for len(getOids) > 0 {
		oids := len(getOids)
		if oids > maxOids {
			oids = maxOids
		}

		level.Debug(logger).Log("msg", "Getting OIDs", "oids", oids)
		getStart := time.Now()
		packet, err := snmp.Get(getOids[:oids])
		if err != nil {
			if err == context.Canceled {
				return results, fmt.Errorf("scrape cancelled after %s (possible timeout) getting target %s",
					time.Since(getInitialStart), snmp.Target)
			}
			return results, fmt.Errorf("error getting target %s: %s", snmp.Target, err)
		}
		level.Debug(logger).Log("msg", "Get of OIDs completed", "oids", oids, "duration_seconds", time.Since(getStart))
		// SNMPv1 will return packet error for unsupported OIDs.
		if packet.Error == gosnmp.NoSuchName && snmp.Version == gosnmp.Version1 {
			level.Debug(logger).Log("msg", "OID not supported by target", "oids", getOids[0])
			getOids = getOids[oids:]
			continue
		}
		// Response received with errors.
		// TODO: "stringify" gosnmp errors instead of showing error code.
		if packet.Error != gosnmp.NoError {
			return results, fmt.Errorf("error reported by target %s: Error Status %d", snmp.Target, packet.Error)
		}
		for _, v := range packet.Variables {
			if v.Type == gosnmp.NoSuchObject || v.Type == gosnmp.NoSuchInstance {
				level.Debug(logger).Log("msg", "OID not supported by target", "oids", v.Name)
				continue
			}
			results.pdus = append(results.pdus, v)
		}
		getOids = getOids[oids:]
	}

	for _, subtree := range newWalk {
		var pdus []gosnmp.SnmpPDU
		level.Debug(logger).Log("msg", "Walking subtree", "oid", subtree)
		walkStart := time.Now()
		if snmp.Version == gosnmp.Version1 {
			pdus, err = snmp.WalkAll(subtree)
		} else {
			pdus, err = snmp.BulkWalkAll(subtree)
		}
		if err != nil {
			if err == context.Canceled {
				return results, fmt.Errorf("scrape canceled after %s (possible timeout) walking target %s",
					time.Since(getInitialStart), snmp.Target)
			}
			return results, fmt.Errorf("error walking target %s: %s", snmp.Target, err)
		}
		level.Debug(logger).Log("msg", "Walk of subtree completed", "oid", subtree, "duration_seconds", time.Since(walkStart))

		results.pdus = append(results.pdus, pdus...)
	}
	return results, nil
}

func configureTarget(g *gosnmp.GoSNMP, target string) error {
	if s := strings.SplitN(target, "://", 2); len(s) == 2 {
		g.Transport = s[0]
		target = s[1]
	}
	g.Target = target
	g.Port = 161
	if host, port, err := net.SplitHostPort(target); err == nil {
		g.Target = host
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("error converting port number to int for target %q: %w", target, err)
		}
		g.Port = uint16(p)
	}
	return nil
}

func filterAllowedIndices(logger log.Logger, filter config.DynamicFilter, pdus []gosnmp.SnmpPDU, allowedList []string, metrics Metrics) []string {
	level.Debug(logger).Log("msg", "Evaluating rule for oid", "oid", filter.Oid)
	for _, pdu := range pdus {
		found := false
		for _, val := range filter.Values {
			snmpval := pduValueAsString(&pdu, "DisplayString", metrics)
			level.Debug(logger).Log("config value", val, "snmp value", snmpval)

			if regexp.MustCompile(val).MatchString(snmpval) {
				found = true
				break
			}
		}
		if found {
			pduArray := strings.Split(pdu.Name, ".")
			index := pduArray[len(pduArray)-1]
			level.Debug(logger).Log("msg", "Caching index", "index", index)
			allowedList = append(allowedList, index)
		}
	}
	return allowedList
}

func updateWalkConfig(walkConfig []string, filter config.DynamicFilter, logger log.Logger) []string {
	newCfg := []string{}
	for _, elem := range walkConfig {
		found := false
		for _, targetOid := range filter.Targets {
			if elem == targetOid {
				level.Debug(logger).Log("msg", "Deleting for walk configuration", "oid", targetOid)
				found = true
				break
			}
		}
		// Oid not found in target,  we walk it.
		if !found {
			newCfg = append(newCfg, elem)
		}
	}
	return newCfg
}

func updateGetConfig(getConfig []string, filter config.DynamicFilter, logger log.Logger) []string {
	newCfg := []string{}
	for _, elem := range getConfig {
		found := false
		for _, targetOid := range filter.Targets {
			if strings.HasPrefix(elem, targetOid) {
				found = true
				break
			}
		}
		// Oid not found in targets, we keep it.
		if !found {
			level.Debug(logger).Log("msg", "Keeping get configuration", "oid", elem)
			newCfg = append(newCfg, elem)
		}
	}
	return newCfg
}

func addAllowedIndices(filter config.DynamicFilter, allowedList []string, logger log.Logger, newCfg []string) []string {
	for _, targetOid := range filter.Targets {
		for _, index := range allowedList {
			level.Debug(logger).Log("msg", "Adding get configuration", "oid", targetOid+"."+index)
			newCfg = append(newCfg, targetOid+"."+index)
		}
	}
	return newCfg
}

type MetricNode struct {
	metric *config.Metric

	children map[int]*MetricNode
}

// Build a tree of metrics from the config, for fast lookup when there's lots of them.
func buildMetricTree(metrics []*config.Metric) *MetricNode {
	metricTree := &MetricNode{children: map[int]*MetricNode{}}
	for _, metric := range metrics {
		head := metricTree
		for _, o := range oidToList(metric.Oid) {
			_, ok := head.children[o]
			if !ok {
				head.children[o] = &MetricNode{children: map[int]*MetricNode{}}
			}
			head = head.children[o]
		}
		head.metric = metric
	}
	return metricTree
}

type Metrics struct {
	SNMPCollectionDuration *prometheus.HistogramVec
	SNMPUnexpectedPduType  prometheus.Counter
	SNMPDuration           prometheus.Histogram
	SNMPPackets            prometheus.Counter
	SNMPRetries            prometheus.Counter
}

type NamedModule struct {
	*config.Module
	name string
}

func NewNamedModule(name string, module *config.Module) *NamedModule {
	return &NamedModule{
		Module: module,
		name:   name,
	}
}

type Collector struct {
	ctx         context.Context
	target      string
	auth        *config.Auth
	authName    string
	modules     []*NamedModule
	logger      log.Logger
	metrics     Metrics
	concurrency int
}

func New(ctx context.Context, target, authName string, auth *config.Auth, modules []*NamedModule, logger log.Logger, metrics Metrics, conc int) *Collector {
	return &Collector{ctx: ctx, target: target, authName: authName, auth: auth, modules: modules, logger: logger, metrics: metrics, concurrency: conc}
}

// Describe implements Prometheus.Collector.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (c Collector) collect(ch chan<- prometheus.Metric, module *NamedModule) {
	logger := log.With(c.logger, "module", module.name)
	start := time.Now()
	results, err := ScrapeTarget(c.ctx, c.target, c.auth, module.Module, logger, c.metrics)
	moduleLabel := prometheus.Labels{"module": module.name}
	if err != nil {
		level.Info(logger).Log("msg", "Error scraping target", "err", err)
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error scraping target", nil, moduleLabel), err)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_walk_duration_seconds", "Time SNMP walk/bulkwalk took.", nil, moduleLabel),
		prometheus.GaugeValue,
		time.Since(start).Seconds())
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_packets_sent", "Packets sent for get, bulkget, and walk; including retries.", nil, moduleLabel),
		prometheus.GaugeValue,
		float64(results.packets))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_packets_retried", "Packets retried for get, bulkget, and walk.", nil, moduleLabel),
		prometheus.GaugeValue,
		float64(results.retries))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_pdus_returned", "PDUs returned from get, bulkget, and walk.", nil, moduleLabel),
		prometheus.GaugeValue,
		float64(len(results.pdus)))
	oidToPdu := make(map[string]gosnmp.SnmpPDU, len(results.pdus))
	for _, pdu := range results.pdus {
		oidToPdu[pdu.Name[1:]] = pdu
	}

	if _, ok := ssmMetricRecords.current[c.target]; !ok {
		ssmMetricRecords.current[c.target] = &ssmMetricRecord{
			collectedMetrics: make(map[string]struct{}),
			hrSWRunPerfCPU:   make(map[string]ssmMetricPerfCPU),
			hrSWRunName:      make(map[string]string),
		}
	} else {
		// reset current data if needed
		ssmMetricRecords.current[c.target].hrSWRunPerfMem = 0
		ssmMetricRecords.current[c.target].hrSWRunPerfCPU = make(map[string]ssmMetricPerfCPU)
		ssmMetricRecords.current[c.target].hrSWRunName = make(map[string]string)
	}
	ssmMetricRecords.current[c.target].mu.Lock()
	defer ssmMetricRecords.current[c.target].mu.Unlock()

	metricTree := buildMetricTree(module.Metrics)
	// Look for metrics that match each pdu.
PduLoop:
	for oid, pdu := range oidToPdu {
		head := metricTree
		oidList := oidToList(oid)
		for i, o := range oidList {
			var ok bool
			head, ok = head.children[o]
			if !ok {
				continue PduLoop
			}
			if head.metric == nil {
				continue
			}

			ssmMetricRecords.current[c.target].collectedMetrics[head.metric.Name] = struct{}{}

			// Found a match.
			switch head.metric.Name {
			case "ssCpuRawUser":
				ssmMetricRecords.current[c.target].ssCPURawUser = getPduValue(&pdu)
			case "ssCpuRawNice":
				ssmMetricRecords.current[c.target].ssCPURawNice = getPduValue(&pdu)
			case "ssCpuRawSystem":
				ssmMetricRecords.current[c.target].ssCPURawSystem = getPduValue(&pdu)
			case "ssCpuRawIdle":
				ssmMetricRecords.current[c.target].ssCPURawIdle = getPduValue(&pdu)
			case "ssCpuRawWait":
				ssmMetricRecords.current[c.target].ssCPURawWait = getPduValue(&pdu)
			case "ssCpuRawKernel":
				ssmMetricRecords.current[c.target].ssCPURawKernel = getPduValue(&pdu)
			case "ssCpuRawInterrupt":
				ssmMetricRecords.current[c.target].ssCPURawInterrupt = getPduValue(&pdu)
			case "ssCpuRawSoftIRQ":
				ssmMetricRecords.current[c.target].ssCPURawSoftIRQ = getPduValue(&pdu)
			case "ssCpuRawSteal":
				ssmMetricRecords.current[c.target].ssCPURawSteal = getPduValue(&pdu)
			case "ssCpuRawGuest":
				ssmMetricRecords.current[c.target].ssCPURawGuest = getPduValue(&pdu)
			case "hrSystemDate":
				ssmMetricRecords.current[c.target].hrSystemDate, _ = parseDateAndTime(&pdu)
			case "hrSWRunPerfMem":
				ssmMetricRecords.current[c.target].hrSWRunPerfMem = ssmMetricRecords.current[c.target].hrSWRunPerfMem + getPduValue(&pdu)
			case "hrSWRunPerfCPU":
				labels := indexesToLabels(oidList[i+1:], head.metric, oidToPdu, c.metrics)
				ssmMetricRecords.current[c.target].hrSWRunPerfCPU[labels["hrSWRunIndex"]] = ssmMetricPerfCPU{
					hrSWRunType: labels["hrSWRunType"],
					value:       getPduValue(&pdu),
				}
			case "hrSWRunName":
				labels := indexesToLabels(oidList[i+1:], head.metric, oidToPdu, c.metrics)
				ssmMetricRecords.current[c.target].hrSWRunName[labels["hrSWRunIndex"]] = string(pdu.Value.([]byte))
			default:
				switch head.metric.Name {
				case "hrMemorySize":
					ssmMetricRecords.current[c.target].hrMemorySize = getPduValue(&pdu)
				}

				samples := pduToSamples(oidList[i+1:], &pdu, head.metric, oidToPdu, c.logger, c.metrics)
				for _, sample := range samples {
					ch <- sample
				}
			}
			break
		}
	}

	samples, err := c.collecSSMCPUMetrics()
	if err != nil {
		samples = append(samples, prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling collecSSMCPUMetrics", nil, nil),
			fmt.Errorf("error for metric %s: %v", nodeCPUAverageName, err)))
	}
	for _, sample := range samples {
		ch <- sample
	}

	samples, err = c.collectSSMMemoryMetrics()
	if err != nil {
		samples = append(samples, prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling collectSSMMemoryMetrics", nil, nil),
			fmt.Errorf("error for metric memory: %v", err)))
	}
	for _, sample := range samples {
		ch <- sample
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_duration_seconds", "Total SNMP time scrape took (walk and processing).", nil, moduleLabel),
		prometheus.GaugeValue,
		time.Since(start).Seconds())
	c.copyHistorySSMMetrics()
}

// Collect implements Prometheus.Collector.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	workerCount := c.concurrency
	if workerCount < 1 {
		workerCount = 1
	}
	workerChan := make(chan *NamedModule)
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range workerChan {
				logger := log.With(c.logger, "module", m.name)
				level.Debug(logger).Log("msg", "Starting scrape")
				start := time.Now()
				c.collect(ch, m)
				duration := time.Since(start).Seconds()
				level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
				c.metrics.SNMPCollectionDuration.WithLabelValues(m.name).Observe(duration)
			}
		}()
	}

	for _, module := range c.modules {
		workerChan <- module
	}
	close(workerChan)
	wg.Wait()
}

func getPduValue(pdu *gosnmp.SnmpPDU) float64 {
	switch pdu.Type {
	case gosnmp.Counter64:
		if *wrapCounters {
			// Wrap by 2^53.
			return float64(gosnmp.ToBigInt(pdu.Value).Uint64() % float64Mantissa)
		}
		return float64(gosnmp.ToBigInt(pdu.Value).Uint64())
	case gosnmp.OpaqueFloat:
		return float64(pdu.Value.(float32))
	case gosnmp.OpaqueDouble:
		return pdu.Value.(float64)
	default:
		return float64(gosnmp.ToBigInt(pdu.Value).Int64())
	}
}

// parseDateAndTime extracts a UNIX timestamp from an RFC 2579 DateAndTime.
func parseDateAndTime(pdu *gosnmp.SnmpPDU) (float64, error) {
	var (
		v   []byte
		tz  *time.Location
		err error
	)
	// DateAndTime should be a slice of bytes.
	switch pduType := pdu.Value.(type) {
	case []byte:
		v = pdu.Value.([]byte)
	default:
		return 0, fmt.Errorf("invalid DateAndTime type %v", pduType)
	}
	pduLength := len(v)
	// DateAndTime can be 8 or 11 bytes depending if the time zone is included.
	switch pduLength {
	case 8:
		// No time zone included, assume UTC.
		tz = time.UTC
	case 11:
		// Extract the timezone from the last 3 bytes.
		locString := fmt.Sprintf("%s%02d%02d", string(v[8]), v[9], v[10])
		loc, err := time.Parse("-0700", locString)
		if err != nil {
			return 0, fmt.Errorf("error parsing location string: %q, error: %s", locString, err)
		}
		tz = loc.Location()
	default:
		return 0, fmt.Errorf("invalid DateAndTime length %v", pduLength)
	}
	if err != nil {
		return 0, fmt.Errorf("unable to parse DateAndTime %q, error: %s", v, err)
	}
	// Build the date from the various fields and time zone.
	t := time.Date(
		int(binary.BigEndian.Uint16(v[0:2])),
		time.Month(v[2]),
		int(v[3]),
		int(v[4]),
		int(v[5]),
		int(v[6]),
		int(v[7])*1e+8,
		tz)
	return float64(t.Unix()), nil
}

func pduToSamples(indexOids []int, pdu *gosnmp.SnmpPDU, metric *config.Metric, oidToPdu map[string]gosnmp.SnmpPDU, logger log.Logger, metrics Metrics) []prometheus.Metric {
	var err error
	// The part of the OID that is the indexes.
	labels := indexesToLabels(indexOids, metric, oidToPdu, metrics)

	value := getPduValue(pdu)

	labelnames := make([]string, 0, len(labels)+1)
	labelvalues := make([]string, 0, len(labels)+1)
	for k, v := range labels {
		labelnames = append(labelnames, k)
		labelvalues = append(labelvalues, v)
	}

	var t prometheus.ValueType
	switch metric.Type {
	case config.MetricTypeCounter:
		t = prometheus.CounterValue
	case config.MetricTypeGauge:
		t = prometheus.GaugeValue
	case config.MetricTypeFloat, config.MetricTypeDouble:
		t = prometheus.GaugeValue
	case config.MetricTypeDateAndTime:
		t = prometheus.GaugeValue
		value, err = parseDateAndTime(pdu)
		if err != nil {
			level.Debug(logger).Log("msg", "Error parsing DateAndTime", "err", err)
			return []prometheus.Metric{}
		}
	case config.MetricTypeEnumAsInfo:
		return enumAsInfo(metric, int(value), labelnames, labelvalues)
	case config.MetricTypeEnumAsStateSet:
		return enumAsStateSet(metric, int(value), labelnames, labelvalues)
	case config.MetricTypeBits:
		return bits(metric, pdu.Value, labelnames, labelvalues)
	default:
		// It's some form of string.
		t = prometheus.GaugeValue
		value = 1.0
		metricType := metric.Type

		if typeMapping, ok := combinedTypeMapping[metricType]; ok {
			// Lookup associated sub type in previous object.
			prevOid := fmt.Sprintf("%s.%s", getPrevOid(metric.Oid), listToOid(indexOids))
			if prevPdu, ok := oidToPdu[prevOid]; ok {
				val := int(getPduValue(&prevPdu))
				if t, ok := typeMapping[val]; ok {
					metricType = t
				} else {
					metricType = "OctetString"
					level.Debug(logger).Log("msg", "Unable to handle type value", "value", val, "oid", prevOid, "metric", metric.Name)
				}
			} else {
				metricType = "OctetString"
				level.Debug(logger).Log("msg", "Unable to find type at oid for metric", "oid", prevOid, "metric", metric.Name)
			}
		}

		if len(metric.RegexpExtracts) > 0 {
			return applyRegexExtracts(metric, pduValueAsString(pdu, metricType, metrics), labelnames, labelvalues, logger)
		}
		// For strings we put the value as a label with the same name as the metric.
		// If the name is already an index, we do not need to set it again.
		if _, ok := labels[metric.Name]; !ok {
			labelnames = append(labelnames, metric.Name)
			labelvalues = append(labelvalues, pduValueAsString(pdu, metricType, metrics))
		}
	}

	if metric.Scale != 0.0 {
		value *= metric.Scale
	}
	value += metric.Offset

	var sample prometheus.Metric
	if isSSMMetrics(metric) {
		samples, err := newSSMConstMetric(metric, t, value, labelnames, labelvalues, nil)
		if err == nil {
			return samples
		}
	} else {
		sample, err = prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help, labelnames, nil),
			t, value, labelvalues...)
	}
	if err != nil {
		sample = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric", nil, nil),
			fmt.Errorf("error for metric %s with labels %v from indexOids %v: %v", metric.Name, labelvalues, indexOids, err))
	}

	return []prometheus.Metric{sample}
}

func applyRegexExtracts(metric *config.Metric, pduValue string, labelnames, labelvalues []string, logger log.Logger) []prometheus.Metric {
	results := []prometheus.Metric{}
	for name, strMetricSlice := range metric.RegexpExtracts {
		for _, strMetric := range strMetricSlice {
			indexes := strMetric.Regex.FindStringSubmatchIndex(pduValue)
			if indexes == nil {
				level.Debug(logger).Log("msg", "No match found for regexp", "metric", metric.Name, "value", pduValue, "regex", strMetric.Regex.String())
				continue
			}
			res := strMetric.Regex.ExpandString([]byte{}, strMetric.Value, pduValue, indexes)
			v, err := strconv.ParseFloat(string(res), 64)
			if err != nil {
				level.Debug(logger).Log("msg", "Error parsing float64 from value", "metric", metric.Name, "value", pduValue, "regex", strMetric.Regex.String(), "extracted_value", res)
				continue
			}
			newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name+name, metric.Help+" (regex extracted)", labelnames, nil),
				prometheus.GaugeValue, v, labelvalues...)
			if err != nil {
				newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for regex_extract", nil, nil),
					fmt.Errorf("error for metric %s with labels %v: %v", metric.Name+name, labelvalues, err))
			}
			results = append(results, newMetric)
			break
		}
	}
	return results
}

func enumAsInfo(metric *config.Metric, value int, labelnames, labelvalues []string) []prometheus.Metric {
	// Lookup enum, default to the value.
	state, ok := metric.EnumValues[int(value)]
	if !ok {
		state = strconv.Itoa(int(value))
	}
	labelnames = append(labelnames, metric.Name)
	labelvalues = append(labelvalues, state)

	newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name+"_info", metric.Help+" (EnumAsInfo)", labelnames, nil),
		prometheus.GaugeValue, 1.0, labelvalues...)
	if err != nil {
		newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for EnumAsInfo", nil, nil),
			fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
	}
	return []prometheus.Metric{newMetric}
}

func enumAsStateSet(metric *config.Metric, value int, labelnames, labelvalues []string) []prometheus.Metric {
	labelnames = append(labelnames, metric.Name)
	results := []prometheus.Metric{}

	state, ok := metric.EnumValues[value]
	if !ok {
		// Fallback to using the value.
		state = strconv.Itoa(value)
	}
	newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help+" (EnumAsStateSet)", labelnames, nil),
		prometheus.GaugeValue, 1.0, append(labelvalues, state)...)
	if err != nil {
		newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for EnumAsStateSet", nil, nil),
			fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
	}
	results = append(results, newMetric)

	for k, v := range metric.EnumValues {
		if k == value {
			continue
		}
		newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help+" (EnumAsStateSet)", labelnames, nil),
			prometheus.GaugeValue, 0.0, append(labelvalues, v)...)
		if err != nil {
			newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for EnumAsStateSet", nil, nil),
				fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
		}
		results = append(results, newMetric)
	}
	return results
}

func bits(metric *config.Metric, value interface{}, labelnames, labelvalues []string) []prometheus.Metric {
	bytes, ok := value.([]byte)
	if !ok {
		return []prometheus.Metric{prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "BITS type was not a BISTRING on the wire.", nil, nil),
			fmt.Errorf("error for metric %s with labels %v: %T", metric.Name, labelvalues, value))}
	}
	labelnames = append(labelnames, metric.Name)
	results := []prometheus.Metric{}

	for k, v := range metric.EnumValues {
		bit := 0.0
		// Most significant byte most significant bit, then most significant byte 2nd most significant bit etc.
		if k < len(bytes)*8 {
			if (bytes[k/8] & (128 >> (k % 8))) != 0 {
				bit = 1.0
			}
		}
		newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help+" (Bits)", labelnames, nil),
			prometheus.GaugeValue, bit, append(labelvalues, v)...)
		if err != nil {
			newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for Bits", nil, nil),
				fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
		}
		results = append(results, newMetric)
	}
	return results
}

// Right pad oid with zeros, and split at the given point.
// Some routers exclude trailing 0s in responses.
func splitOid(oid []int, count int) ([]int, []int) {
	head := make([]int, count)
	tail := []int{}
	for i, v := range oid {
		if i < count {
			head[i] = v
		} else {
			tail = append(tail, v)
		}
	}
	return head, tail
}

// This mirrors decodeValue in gosnmp's helper.go.
func pduValueAsString(pdu *gosnmp.SnmpPDU, typ string, metrics Metrics) string {
	switch pdu.Value.(type) {
	case int:
		return strconv.Itoa(pdu.Value.(int))
	case uint:
		return strconv.FormatUint(uint64(pdu.Value.(uint)), 10)
	case uint64:
		return strconv.FormatUint(pdu.Value.(uint64), 10)
	case float32:
		return strconv.FormatFloat(float64(pdu.Value.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(pdu.Value.(float64), 'f', -1, 64)
	case string:
		if pdu.Type == gosnmp.ObjectIdentifier {
			// Trim leading period.
			return pdu.Value.(string)[1:]
		}
		// DisplayString.
		return strings.ToValidUTF8(pdu.Value.(string), "�")
	case []byte:
		if typ == "" || typ == "Bits" {
			typ = "OctetString"
		}
		// Reuse the OID index parsing code.
		parts := make([]int, len(pdu.Value.([]byte)))
		for i, o := range pdu.Value.([]byte) {
			parts[i] = int(o)
		}
		if typ == "OctetString" || typ == "DisplayString" {
			// Prepend the length, as it is explicit in an index.
			parts = append([]int{len(pdu.Value.([]byte))}, parts...)
		}
		str, _, _ := indexOidsAsString(parts, typ, 0, false, nil)
		return strings.ToValidUTF8(str, "�")
	case nil:
		return ""
	default:
		// This shouldn't happen.
		metrics.SNMPUnexpectedPduType.Inc()
		return fmt.Sprintf("%s", pdu.Value)
	}
}

// Convert oids to a string index value.
//
// Returns the string, the oids that were used and the oids left over.
func indexOidsAsString(indexOids []int, typ string, fixedSize int, implied bool, enumValues map[int]string) (string, []int, []int) {
	if typeMapping, ok := combinedTypeMapping[typ]; ok {
		subOid, valueOids := splitOid(indexOids, 2)
		if typ == "InetAddressMissingSize" {
			// The size of the main index value is missing.
			subOid, valueOids = splitOid(indexOids, 1)
		}
		var str string
		var used, remaining []int
		if t, ok := typeMapping[subOid[0]]; ok {
			str, used, remaining = indexOidsAsString(valueOids, t, 0, false, enumValues)
			return str, append(subOid, used...), remaining
		}
		if typ == "InetAddressMissingSize" {
			// We don't know the size, so pass everything remaining.
			return indexOidsAsString(indexOids, "OctetString", 0, true, enumValues)
		}
		// The 2nd oid is the length.
		return indexOidsAsString(indexOids, "OctetString", subOid[1]+2, false, enumValues)
	}

	switch typ {
	case "Integer32", "Integer", "gauge", "counter":
		// Extract the oid for this index, and keep the remainder for the next index.
		subOid, indexOids := splitOid(indexOids, 1)
		return fmt.Sprintf("%d", subOid[0]), subOid, indexOids
	case "PhysAddress48":
		subOid, indexOids := splitOid(indexOids, 6)
		parts := make([]string, 6)
		for i, o := range subOid {
			parts[i] = fmt.Sprintf("%02X", o)
		}
		return strings.Join(parts, ":"), subOid, indexOids
	case "OctetString":
		var subOid []int
		// The length of fixed size indexes come from the MIB.
		// For varying size, we read it from the first oid.
		length := fixedSize
		if implied {
			length = len(indexOids)
		}
		if length == 0 {
			subOid, indexOids = splitOid(indexOids, 1)
			length = subOid[0]
		}
		content, indexOids := splitOid(indexOids, length)
		subOid = append(subOid, content...)
		parts := make([]byte, length)
		for i, o := range content {
			parts[i] = byte(o)
		}
		if len(parts) == 0 {
			return "", subOid, indexOids
		}
		return fmt.Sprintf("0x%X", string(parts)), subOid, indexOids
	case "DisplayString":
		var subOid []int
		length := fixedSize
		if implied {
			length = len(indexOids)
		}
		if length == 0 {
			subOid, indexOids = splitOid(indexOids, 1)
			length = subOid[0]
		}
		content, indexOids := splitOid(indexOids, length)
		subOid = append(subOid, content...)
		parts := make([]byte, length)
		for i, o := range content {
			parts[i] = byte(o)
		}
		// ASCII, so can convert staight to utf-8.
		return string(parts), subOid, indexOids
	case "InetAddressIPv4":
		subOid, indexOids := splitOid(indexOids, 4)
		parts := make([]string, 4)
		for i, o := range subOid {
			parts[i] = strconv.Itoa(o)
		}
		return strings.Join(parts, "."), subOid, indexOids
	case "InetAddressIPv6":
		subOid, indexOids := splitOid(indexOids, 16)
		parts := make([]interface{}, 16)
		for i, o := range subOid {
			parts[i] = o
		}
		return fmt.Sprintf("%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X", parts...), subOid, indexOids
	case "EnumAsInfo":
		subOid, indexOids := splitOid(indexOids, 1)
		value, ok := enumValues[subOid[0]]
		if ok {
			return value, subOid, indexOids
		}
		return fmt.Sprintf("%d", subOid[0]), subOid, indexOids
	default:
		panic(fmt.Sprintf("Unknown index type %s", typ))
	}
}

func getPrevOid(oid string) string {
	oids := strings.Split(oid, ".")
	i, _ := strconv.Atoi(oids[len(oids)-1])
	oids[len(oids)-1] = strconv.Itoa(i - 1)
	return strings.Join(oids, ".")
}

func indexesToLabels(indexOids []int, metric *config.Metric, oidToPdu map[string]gosnmp.SnmpPDU, metrics Metrics) map[string]string {
	labels := map[string]string{}
	labelOids := map[string][]int{}

	// Covert indexes to useful strings.
	for _, index := range metric.Indexes {
		str, subOid, remainingOids := indexOidsAsString(indexOids, index.Type, index.FixedSize, index.Implied, index.EnumValues)
		// The labelvalue is the text form of the index oids.
		labels[index.Labelname] = str
		// Save its oid in case we need it for lookups.
		labelOids[index.Labelname] = subOid
		// For the next iteration.
		indexOids = remainingOids
	}

	// Perform lookups.
	for _, lookup := range metric.Lookups {
		if len(lookup.Labels) == 0 {
			delete(labels, lookup.Labelname)
			continue
		}
		oid := lookup.Oid
		for _, label := range lookup.Labels {
			oid = fmt.Sprintf("%s.%s", oid, listToOid(labelOids[label]))
		}
		if pdu, ok := oidToPdu[oid]; ok {
			t := lookup.Type
			if typeMapping, ok := combinedTypeMapping[lookup.Type]; ok {
				// Lookup associated sub type in previous object.
				prevOid := getPrevOid(lookup.Oid)
				for _, label := range lookup.Labels {
					prevOid = fmt.Sprintf("%s.%s", prevOid, listToOid(labelOids[label]))
				}
				if prevPdu, ok := oidToPdu[prevOid]; ok {
					val := int(getPduValue(&prevPdu))
					if ty, ok := typeMapping[val]; ok {
						t = ty
					}
				}
			}
			labels[lookup.Labelname] = pduValueAsString(&pdu, t, metrics)
			labelOids[lookup.Labelname] = []int{int(gosnmp.ToBigInt(pdu.Value).Int64())}
		} else {
			labels[lookup.Labelname] = ""
		}
	}

	return labels
}
