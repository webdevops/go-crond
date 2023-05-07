package main

import "github.com/prometheus/client_golang/prometheus"

var (
	prometheusMetricTask            *prometheus.GaugeVec
	prometheusMetricTaskRunCount    *prometheus.CounterVec
	prometheusMetricTaskRunResult   *prometheus.GaugeVec
	prometheusMetricTaskRunTime     *prometheus.GaugeVec
	prometheusMetricTaskRunPrevTs   *prometheus.GaugeVec
	prometheusMetricTaskRunNextTs   *prometheus.GaugeVec
	prometheusMetricTaskRunDuration *prometheus.GaugeVec
)

func initMetrics() {
	prometheusMetricTask = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_info",
			Help: "gocrond task info",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(prometheusMetricTask)

	prometheusMetricTaskRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gocrond_task_run_count",
			Help: "gocrond task run count",
		},
		[]string{"cronSpec", "cronUser", "cronCommand", "result"},
	)
	prometheus.MustRegister(prometheusMetricTaskRunCount)

	prometheusMetricTaskRunResult = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_result",
			Help: "gocrond task run result",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(prometheusMetricTaskRunResult)

	prometheusMetricTaskRunTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_time",
			Help: "gocrond task last run time",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(prometheusMetricTaskRunTime)

	prometheusMetricTaskRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_duration",
			Help: "gocrond task last run duration",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(prometheusMetricTaskRunDuration)

	prometheusMetricTaskRunNextTs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_next_time",
			Help: "gocrond task next run ts",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(prometheusMetricTaskRunNextTs)

	prometheusMetricTaskRunPrevTs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_prev_time",
			Help: "gocrond task prev run ts",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(prometheusMetricTaskRunPrevTs)
}

func resetMetrics() {
	prometheusMetricTask.Reset()
	prometheusMetricTaskRunCount.Reset()
	prometheusMetricTaskRunResult.Reset()
	prometheusMetricTaskRunTime.Reset()
	prometheusMetricTaskRunDuration.Reset()
	prometheusMetricTaskRunNextTs.Reset()
	prometheusMetricTaskRunPrevTs.Reset()
}
