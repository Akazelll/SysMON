classDiagram
    %% Layer Models
    class SystemMetric {
        +float64 CPUUsage
        +float64 RAMUsage
        +float64 DiskUsage
        +float64 NetRXSpeed
        +float64 NetTXSpeed
    }
    class ProcessStat {
        +int32 PID
        +string Name
        +float64 CPUUsage
        +float32 RAMUsage
    }
    class AlertRule {
        +string Metric
        +float64 Threshold
        +int Duration
    }
    class RingBuffer {
        -[]float64 Data
        -int size
        +Add(value: float64)
    }

    %% Layer Repository
    class OSMonitor {
        -net.IOCountersStat lastNetStat
        -time.Time lastTime
        +GetCurrentMetrics() SystemMetric
        +GetTopProcesses() []ProcessStat
    }

    %% Layer Services
    class AlertEngine {
        +[]AlertRule Rules
        +fyne.App App
        -map breachCounters
        +AddRule(rule: AlertRule)
        +Evaluate(metrics: SystemMetric)
    }
    class Exporter {
        +ExportCPUHistory(writer: io.Writer, cpuData: []float64) error
    }

    %% Layer UI (Presentation)
    class LineChart {
        +fyne.Container Container
        +color.Color Color
        +Update(data: []float64, width: float32, height: float32)
    }
    class ProcessTable {
        +widget.Table Table
        +[]ProcessStat Data
        +UpdateData(newData: []ProcessStat)
    }

    %% Relasi
    OSMonitor --> SystemMetric : creates
    OSMonitor --> ProcessStat : creates
    AlertEngine --> AlertRule : contains
    AlertEngine ..> SystemMetric : evaluates
    ProcessTable ..> ProcessStat : displays
    LineChart ..> RingBuffer : visualizes