package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const namespace = "nvidia"

var usePerProcess = false
var prevProcesses []*Process

type Exporter struct {
	up                        prometheus.Gauge
	info                      *prometheus.GaugeVec
	deviceCount               prometheus.Gauge
	temperatures              *prometheus.GaugeVec
	deviceInfo                *prometheus.GaugeVec
	powerUsage                *prometheus.GaugeVec
	powerLimit                *prometheus.GaugeVec
	fanSpeed                  *prometheus.GaugeVec
	memoryTotal               *prometheus.GaugeVec
	memoryUsed                *prometheus.GaugeVec
	utilizationMemory         *prometheus.GaugeVec
	utilizationGPU            *prometheus.GaugeVec
	clockCurrentGraphics      *prometheus.GaugeVec
	clockCurrentMemory        *prometheus.GaugeVec
	utilizationProcessName    *prometheus.GaugeVec
	utilizationProcessSMUtil  *prometheus.GaugeVec
	utilizationProcessMemUtil *prometheus.GaugeVec
	utilizationProcessEncUtil *prometheus.GaugeVec
	utilizationProcessDecUtil *prometheus.GaugeVec
	pcieTxBytes               *prometheus.GaugeVec
	pcieRxBytes               *prometheus.GaugeVec
}

func main() {
	var (
		level         = flag.String("log.level", "info", "Set the output log level")
		listenAddress = flag.String("web.listen-address", "0.0.0.0:9401", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.BoolVar(&usePerProcess, "nvidia.per-process", false, "Export per-process utilization")
	flag.Parse()
	setLogLevel(*level)

	prometheus.MustRegister(NewExporter())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>NVML Exporter</title></head>
             <body>
             <h1>NVML Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
	     <h2>More information:</h2>
	     <p><a href="https://github.com/cosandr/nvidia-exporter">github.com/cosandr/nvidia-exporter</a></p>
             </body>
             </html>`))
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	log.Infof("Starting HTTP server on %s", *listenAddress)
	log.Infof("Export per-process utilization? %t", usePerProcess)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func setLogLevel(level string) {
	switch level {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.Warnln("Unrecognized minimum log level; using 'info' as default")
	}
}

func NewExporter() *Exporter {
	return &Exporter{
		up: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "up",
				Help:      "NVML Metric Collection Operational",
			},
		),
		info: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "driver_info",
				Help:      "NVML Info",
			},
			[]string{"version"},
		),
		deviceCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "device_count",
				Help:      "Count of found nvidia devices",
			},
		),
		deviceInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "info",
				Help:      "Info as reported by the device",
			},
			[]string{"index", "minor", "uuid", "name"},
		),
		temperatures: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "temperatures",
				Help:      "Temperature as reported by the device",
			},
			[]string{"minor"},
		),
		powerUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "power_usage",
				Help:      "Power usage as reported by the device",
			},
			[]string{"minor"},
		),
		powerLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "power_limit",
				Help:      "Power limit as reported by the device in mW",
			},
			[]string{"minor"},
		),
		fanSpeed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "fanspeed",
				Help:      "Fan speed as reported by the device",
			},
			[]string{"minor"},
		),
		memoryTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_total",
				Help:      "Total memory as reported by the device",
			},
			[]string{"minor"},
		),
		memoryUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_used",
				Help:      "Used memory as reported by the device",
			},
			[]string{"minor"},
		),
		utilizationMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_memory",
				Help:      "Memory Utilization as reported by the device",
			},
			[]string{"minor"},
		),
		utilizationGPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_gpu",
				Help:      "GPU utilization as reported by the device",
			},
			[]string{"minor"},
		),
		clockCurrentGraphics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "clock_current_graphics",
				Help:      "Current GPU graphics clock speed as reported by the device",
			},
			[]string{"minor"},
		),
		clockCurrentMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "clock_current_memory",
				Help:      "Current GPU memory clock speed as reported by the device",
			},
			[]string{"minor"},
		),
		utilizationProcessName: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_process_name",
				Help:      "Process name, if value is 0 the name couldn't be determined",
			},
			[]string{"minor", "pid", "name"},
		),
		utilizationProcessSMUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_process_smutil",
				Help:      "Process SM utilization stats averaged over 10s",
			},
			[]string{"minor", "pid"},
		),
		utilizationProcessMemUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_process_memutil",
				Help:      "Process memory utilization stats averaged over 10s",
			},
			[]string{"minor", "pid"},
		),
		utilizationProcessEncUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_process_encutil",
				Help:      "Process encoder utilization stats averaged over 10s",
			},
			[]string{"minor", "pid"},
		),
		utilizationProcessDecUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "utilization_process_decutil",
				Help:      "Process decoder utilization stats averaged over 10s",
			},
			[]string{"minor", "pid"},
		),
		pcieTxBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pcie_tx_bytes",
				Help:      "PCIe TX throughput as reported by the device",
			},
			[]string{"minor"},
		),
		pcieRxBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pcie_rx_bytes",
				Help:      "PCIe RX throughput as reported by the device",
			},
			[]string{"minor"},
		),
	}
}

// This function is used to check if metric
// value is valid; we expect nothing less than 0
// gonvml returns uint data type
func checkMetric(value float64) bool {
	if value < 0 {
		return false
	} else {
		return true
	}
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	data, err := collectMetrics()
	if err != nil {
		log.Errorf("Failed to collect metrics: %s", err)
		e.up.Set(0)
		e.up.Collect(metrics)
		return
	}

	e.up.Set(1)
	e.info.WithLabelValues(data.Version).Set(1)
	e.deviceCount.Set(float64(len(data.Devices)))

	for i := range data.Devices {
		d := data.Devices[i]
		e.deviceInfo.WithLabelValues(d.Index, d.MinorNumber, d.Name, d.UUID).Set(1)
		if checkMetric(d.FanSpeed) {
			e.fanSpeed.WithLabelValues(d.MinorNumber).Set(d.FanSpeed)
		}
		if checkMetric(d.MemoryTotal) {
			e.memoryTotal.WithLabelValues(d.MinorNumber).Set(d.MemoryTotal)
		}
		if checkMetric(d.MemoryUsed) {
			e.memoryUsed.WithLabelValues(d.MinorNumber).Set(d.MemoryUsed)
		}
		if checkMetric(d.PowerUsage) {
			e.powerUsage.WithLabelValues(d.MinorNumber).Set(d.PowerUsage)
		}
		if checkMetric(d.PowerLimit) {
			e.powerLimit.WithLabelValues(d.MinorNumber).Set(d.PowerLimit)
		}
		if checkMetric(d.Temperature) {
			e.temperatures.WithLabelValues(d.MinorNumber).Set(d.Temperature)
		}
		if checkMetric(d.UtilizationGPU) {
			e.utilizationGPU.WithLabelValues(d.MinorNumber).Set(d.UtilizationGPU)
		}
		if checkMetric(d.UtilizationMemory) {
			e.utilizationMemory.WithLabelValues(d.MinorNumber).Set(d.UtilizationMemory)
		}
		if checkMetric(d.ClockCurrentGraphics) {
			e.clockCurrentGraphics.WithLabelValues(d.MinorNumber).Set(d.ClockCurrentGraphics)
		}
		if checkMetric(d.ClockCurrentMemory) {
			e.clockCurrentMemory.WithLabelValues(d.MinorNumber).Set(d.ClockCurrentMemory)
		}
		// Delete previous process metrics
		if len(prevProcesses) > 0 {
			for _, p := range prevProcesses {
				if p.Name != nil {
					e.utilizationProcessName.DeleteLabelValues(d.MinorNumber, p.PromPID(), *p.Name)
				} else {
					e.utilizationProcessName.DeleteLabelValues(d.MinorNumber, p.PromPID(), "N/A")
				}
				e.utilizationProcessSMUtil.DeleteLabelValues(d.MinorNumber, p.PromPID())
				e.utilizationProcessMemUtil.DeleteLabelValues(d.MinorNumber, p.PromPID())
				e.utilizationProcessEncUtil.DeleteLabelValues(d.MinorNumber, p.PromPID())
				e.utilizationProcessDecUtil.DeleteLabelValues(d.MinorNumber, p.PromPID())
				prevProcesses = nil
			}
		}
		if len(d.UtilizationProcesses) > 0 {
			for _, p := range d.UtilizationProcesses {
				if p.Name != nil {
					e.utilizationProcessName.WithLabelValues(d.MinorNumber, p.PromPID(), *p.Name).Set(1)
				} else {
					e.utilizationProcessName.WithLabelValues(d.MinorNumber, p.PromPID(), "N/A").Set(0)
				}
				e.utilizationProcessSMUtil.WithLabelValues(d.MinorNumber, p.PromPID()).Set(float64(p.SMUtil))
				e.utilizationProcessMemUtil.WithLabelValues(d.MinorNumber, p.PromPID()).Set(float64(p.MemUtil))
				e.utilizationProcessEncUtil.WithLabelValues(d.MinorNumber, p.PromPID()).Set(float64(p.EncUtil))
				e.utilizationProcessDecUtil.WithLabelValues(d.MinorNumber, p.PromPID()).Set(float64(p.DecUtil))
				prevProcesses = append(prevProcesses, p)
			}
		}
		if checkMetric(d.PcieTxBytes) {
			e.pcieTxBytes.WithLabelValues(d.MinorNumber).Set(d.PcieTxBytes)
		}
		if checkMetric(d.PcieRxBytes) {
			e.pcieRxBytes.WithLabelValues(d.MinorNumber).Set(d.PcieRxBytes)
		}
	}

	e.deviceCount.Collect(metrics)
	e.deviceInfo.Collect(metrics)
	e.fanSpeed.Collect(metrics)
	e.info.Collect(metrics)
	e.memoryTotal.Collect(metrics)
	e.memoryUsed.Collect(metrics)
	e.powerUsage.Collect(metrics)
	e.powerLimit.Collect(metrics)
	e.temperatures.Collect(metrics)
	e.up.Collect(metrics)
	e.utilizationGPU.Collect(metrics)
	e.utilizationMemory.Collect(metrics)
	e.clockCurrentGraphics.Collect(metrics)
	e.clockCurrentMemory.Collect(metrics)
	if usePerProcess {
		e.utilizationProcessName.Collect(metrics)
		e.utilizationProcessSMUtil.Collect(metrics)
		e.utilizationProcessMemUtil.Collect(metrics)
		e.utilizationProcessEncUtil.Collect(metrics)
		e.utilizationProcessDecUtil.Collect(metrics)
	}
	e.pcieTxBytes.Collect(metrics)
	e.pcieRxBytes.Collect(metrics)
}

func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	e.deviceCount.Describe(descs)
	e.deviceInfo.Describe(descs)
	e.fanSpeed.Describe(descs)
	e.info.Describe(descs)
	e.memoryTotal.Describe(descs)
	e.memoryUsed.Describe(descs)
	e.powerUsage.Describe(descs)
	e.powerLimit.Describe(descs)
	e.temperatures.Describe(descs)
	e.up.Describe(descs)
	e.utilizationGPU.Describe(descs)
	e.utilizationMemory.Describe(descs)
	e.clockCurrentGraphics.Describe(descs)
	e.clockCurrentMemory.Describe(descs)
	if usePerProcess {
		e.utilizationProcessName.Describe(descs)
		e.utilizationProcessSMUtil.Describe(descs)
		e.utilizationProcessMemUtil.Describe(descs)
		e.utilizationProcessEncUtil.Describe(descs)
		e.utilizationProcessDecUtil.Describe(descs)
	}
	e.pcieTxBytes.Describe(descs)
	e.pcieRxBytes.Describe(descs)
}
