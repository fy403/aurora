package metrics

import (
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

// 定义系统metrics指标
type SysExporter struct {
	serviceGaugeVec      prometheus.GaugeVec // 服务信息
	goroutineNumGaugeVec prometheus.GaugeVec // goroutine数量
	memGaugeVec          prometheus.GaugeVec // 系统内存
	cpuGaugeVec          prometheus.GaugeVec // 系统cpu
	gcNumGaugeVec        prometheus.GaugeVec // gc次数
	gcPauseTotalGaugeVec prometheus.GaugeVec // 所有暂停收集垃圾消耗的总时间
	fdsNumGaugeVec       prometheus.GaugeVec // 进程fds
	procCpuGaugeVec      prometheus.GaugeVec // 进程cpu
	procMemGaugeVec      prometheus.GaugeVec // 进程mem
	region               string
	hostname             string
	version              string
	proc                 *process.Process // 进程处理
}

func NewSysExporter(region, version, hostname string, metricsConfig MetricsConfig) *SysExporter {
	serviceGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{ // 带labels的需要使用指针
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "service_info",
		Help:      "服务信息",
	}, []string{"region", "hostname", "version"})

	goroutineNumGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "goroutine_num_total",
		Help:      "goroutine实时数量",
	}, []string{"region", "hostname", "version"})

	gcNumGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "gc_num_total",
		Help:      "gc总次数",
	}, []string{"region", "hostname", "version"})

	gcPauseTotalGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "gc_pause_total",
		Help:      "gc总耗时",
	}, []string{"region", "hostname", "version"})

	fdsNumGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "fds_num_total",
		Help:      "打开的文件描述符",
	}, []string{"region", "hostname", "version"})

	cpuGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "cpu_status_total",
		Help:      "系统cpu使用(%)",
	}, []string{"region", "hostname", "version", "status"})

	memGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "memory_status_total",
		Help:      "系统内存使用统计(M)",
	}, []string{"region", "hostname", "version", "status"})

	procCpuGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "proc_cpu_usage",
		Help:      "进程cpu使用率(%)",
	}, []string{"region", "hostname", "version"})

	procMemGaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsConfig.Namespace,
		Subsystem: metricsConfig.Subsystem,
		Name:      "proc_memory_usage",
		Help:      "进程mem使用率(%)",
	}, []string{"region", "hostname", "version"})

	proc, _ := process.NewProcess(int32(os.Getpid()))
	return &SysExporter{
		serviceGaugeVec:      serviceGaugeVec,
		goroutineNumGaugeVec: goroutineNumGaugeVec,
		memGaugeVec:          memGaugeVec,
		gcNumGaugeVec:        gcNumGaugeVec,
		gcPauseTotalGaugeVec: gcPauseTotalGaugeVec,
		fdsNumGaugeVec:       fdsNumGaugeVec,
		cpuGaugeVec:          cpuGaugeVec,
		procCpuGaugeVec:      procCpuGaugeVec,
		procMemGaugeVec:      procMemGaugeVec,
		region:               region,
		version:              version,
		hostname:             hostname,
		proc:                 proc,
	}
}

// 系统指标获取业务逻辑, 重写Collect(注意：不需要计时器)
func (e *SysExporter) Collect(ch chan<- prometheus.Metric) {
	// service info
	e.serviceGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(float64(1))
	e.serviceGaugeVec.Collect(ch)

	// goroutine数量
	goroutineNum := runtime.NumGoroutine()
	//log.Printf("current get goroutineNum: %d\n", goroutineNum)
	e.goroutineNumGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(float64(goroutineNum))
	e.goroutineNumGaugeVec.Collect(ch)

	// 内存统计
	vm, _ := mem.VirtualMemory()
	e.memGaugeVec.WithLabelValues(e.region, e.hostname, e.version, "total").Set(float64(vm.Total) / 1024 / 1024)    // total_mem(M)
	e.memGaugeVec.WithLabelValues(e.region, e.hostname, e.version, "free").Set(float64(vm.Available) / 1024 / 1024) // free_mem(M)
	e.memGaugeVec.Collect(ch)

	// gc
	gcStats := &debug.GCStats{PauseQuantiles: make([]time.Duration, 5)}
	debug.ReadGCStats(gcStats)
	e.gcNumGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(float64(gcStats.NumGC))
	e.gcNumGaugeVec.Collect(ch)
	e.gcPauseTotalGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(float64(gcStats.PauseTotal))
	e.gcPauseTotalGaugeVec.Collect(ch)

	// 系统cpu
	stats, _ := cpu.Times(false)
	if len(stats) > 0 {
		stat := stats[0]
		idle := stat.Idle / stat.Total()
		usage := 1 - idle
		e.cpuGaugeVec.WithLabelValues(e.region, e.hostname, e.version, "idle").Set(float64(idle))
		e.cpuGaugeVec.WithLabelValues(e.region, e.hostname, e.version, "usage").Set(float64(usage))
		e.cpuGaugeVec.Collect(ch)
	}

	// 进程fd
	fdsNum, _ := e.proc.NumFDs()
	e.fdsNumGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(float64(fdsNum))
	e.fdsNumGaugeVec.Collect(ch)

	// 进程cpu(%)
	pcPercent, _ := e.proc.CPUPercent()
	e.procCpuGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(pcPercent)
	e.procCpuGaugeVec.Collect(ch)

	// 进程mem(%)
	pmPercent, _ := e.proc.MemoryPercent()
	e.procMemGaugeVec.WithLabelValues(e.region, e.hostname, e.version).Set(float64(pmPercent))
	e.procMemGaugeVec.Collect(ch)
}

// 重写metric描述
func (e *SysExporter) Describe(ch chan<- *prometheus.Desc) {
	e.serviceGaugeVec.Describe(ch)
	e.goroutineNumGaugeVec.Describe(ch)
	e.memGaugeVec.Describe(ch)
	e.gcNumGaugeVec.Describe(ch)
	e.gcPauseTotalGaugeVec.Describe(ch)
	e.fdsNumGaugeVec.Describe(ch)
	e.cpuGaugeVec.Describe(ch)
	e.procCpuGaugeVec.Describe(ch)
	e.procMemGaugeVec.Describe(ch)
}
