// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/ruler/manager_metrics.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package ruler

import (
	"github.com/go-kit/log"
	dskit_metrics "github.com/grafana/dskit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// ManagerMetrics aggregates metrics exported by the Prometheus
// rules package and returns them as Mimir metrics
type ManagerMetrics struct {
	regs *dskit_metrics.TenantRegistries

	EvalDuration         *prometheus.Desc
	IterationDuration    *prometheus.Desc
	IterationsMissed     *prometheus.Desc
	IterationsScheduled  *prometheus.Desc
	EvalTotal            *prometheus.Desc
	EvalFailures         *prometheus.Desc
	GroupInterval        *prometheus.Desc
	GroupLastEvalTime    *prometheus.Desc
	GroupLastDuration    *prometheus.Desc
	GroupRules           *prometheus.Desc
	GroupLastEvalSamples *prometheus.Desc
}

// NewManagerMetrics returns a ManagerMetrics struct
func NewManagerMetrics() *ManagerMetrics {
	return &ManagerMetrics{
		regs: dskit_metrics.NewTenantRegistries(log.NewNopLogger()),

		EvalDuration: prometheus.NewDesc(
			"cortex_prometheus_rule_evaluation_duration_seconds",
			"The duration for a rule to execute.",
			[]string{"user"},
			nil,
		),
		IterationDuration: prometheus.NewDesc(
			"cortex_prometheus_rule_group_duration_seconds",
			"The duration of rule group evaluations.",
			[]string{"user"},
			nil,
		),
		IterationsMissed: prometheus.NewDesc(
			"cortex_prometheus_rule_group_iterations_missed_total",
			"The total number of rule group evaluations missed due to slow rule group evaluation.",
			[]string{"user", "rule_group"},
			nil,
		),
		IterationsScheduled: prometheus.NewDesc(
			"cortex_prometheus_rule_group_iterations_total",
			"The total number of scheduled rule group evaluations, whether executed or missed.",
			[]string{"user", "rule_group"},
			nil,
		),
		EvalTotal: prometheus.NewDesc(
			"cortex_prometheus_rule_evaluations_total",
			"The total number of rule evaluations.",
			[]string{"user", "rule_group"},
			nil,
		),
		EvalFailures: prometheus.NewDesc(
			"cortex_prometheus_rule_evaluation_failures_total",
			"The total number of rule evaluation failures.",
			[]string{"user", "rule_group"},
			nil,
		),
		GroupInterval: prometheus.NewDesc(
			"cortex_prometheus_rule_group_interval_seconds",
			"The interval of a rule group.",
			[]string{"user", "rule_group"},
			nil,
		),
		GroupLastEvalTime: prometheus.NewDesc(
			"cortex_prometheus_rule_group_last_evaluation_timestamp_seconds",
			"The timestamp of the last rule group evaluation in seconds.",
			[]string{"user", "rule_group"},
			nil,
		),
		GroupLastDuration: prometheus.NewDesc(
			"cortex_prometheus_rule_group_last_duration_seconds",
			"The duration of the last rule group evaluation.",
			[]string{"user", "rule_group"},
			nil,
		),
		GroupRules: prometheus.NewDesc(
			"cortex_prometheus_rule_group_rules",
			"The number of rules.",
			[]string{"user", "rule_group"},
			nil,
		),
		GroupLastEvalSamples: prometheus.NewDesc(
			"cortex_prometheus_last_evaluation_samples",
			"The number of samples returned during the last rule group evaluation.",
			[]string{"user", "rule_group"},
			nil,
		),
	}
}

// AddTenantRegistry adds a user-specific Prometheus registry.
func (m *ManagerMetrics) AddTenantRegistry(user string, reg *prometheus.Registry) {
	m.regs.AddTenantRegistry(user, reg)
}

// RemoveTenantRegistry removes user-specific Prometheus registry.
func (m *ManagerMetrics) RemoveTenantRegistry(user string) {
	m.regs.RemoveTenantRegistry(user, true)
}

// Describe implements the Collector interface
func (m *ManagerMetrics) Describe(out chan<- *prometheus.Desc) {
	out <- m.EvalDuration
	out <- m.IterationDuration
	out <- m.IterationsMissed
	out <- m.IterationsScheduled
	out <- m.EvalTotal
	out <- m.EvalFailures
	out <- m.GroupInterval
	out <- m.GroupLastEvalTime
	out <- m.GroupLastDuration
	out <- m.GroupRules
	out <- m.GroupLastEvalSamples
}

// Collect implements the Collector interface
func (m *ManagerMetrics) Collect(out chan<- prometheus.Metric) {
	data := m.regs.BuildMetricFamiliesPerTenant()

	// WARNING: It is important that all metrics generated in this method are "Per User".
	// Thanks to that we can actually *remove* metrics for given user (see RemoveTenantRegistry).
	// If same user is later re-added, all metrics will start from 0, which is fine.

	data.SendSumOfSummariesPerTenant(out, m.EvalDuration, "prometheus_rule_evaluation_duration_seconds")
	data.SendSumOfSummariesPerTenant(out, m.IterationDuration, "prometheus_rule_group_duration_seconds")

	data.SendSumOfCountersPerTenant(out, m.IterationsMissed, "prometheus_rule_group_iterations_missed_total", dskit_metrics.WithLabels("rule_group"))
	data.SendSumOfCountersPerTenant(out, m.IterationsScheduled, "prometheus_rule_group_iterations_total", dskit_metrics.WithLabels("rule_group"))
	data.SendSumOfCountersPerTenant(out, m.EvalTotal, "prometheus_rule_evaluations_total", dskit_metrics.WithLabels("rule_group"))
	data.SendSumOfCountersPerTenant(out, m.EvalFailures, "prometheus_rule_evaluation_failures_total", dskit_metrics.WithLabels("rule_group"))
	data.SendSumOfGaugesPerTenantWithLabels(out, m.GroupInterval, "prometheus_rule_group_interval_seconds", "rule_group")
	data.SendSumOfGaugesPerTenantWithLabels(out, m.GroupLastEvalTime, "prometheus_rule_group_last_evaluation_timestamp_seconds", "rule_group")
	data.SendSumOfGaugesPerTenantWithLabels(out, m.GroupLastDuration, "prometheus_rule_group_last_duration_seconds", "rule_group")
	data.SendSumOfGaugesPerTenantWithLabels(out, m.GroupRules, "prometheus_rule_group_rules", "rule_group")
	data.SendSumOfGaugesPerTenantWithLabels(out, m.GroupLastEvalSamples, "prometheus_rule_group_last_evaluation_samples", "rule_group")
}
