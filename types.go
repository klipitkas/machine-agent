package main

import "fmt"

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), []string{"KB", "MB", "GB", "TB", "PB"}[exp])
}

type MachineInfo struct {
	Host        *HostInfo       `json:"host,omitempty"`
	CPU         *CPUInfo        `json:"cpu,omitempty"`
	Memory      *MemoryInfo     `json:"memory,omitempty"`
	Swap        *SwapInfo       `json:"swap,omitempty"`
	Disks       []DiskInfo      `json:"disks,omitempty"`
	Network     *NetworkInfo    `json:"network,omitempty"`
	Load        *LoadInfo       `json:"load,omitempty"`
	Processes   []ProcessInfo   `json:"processes,omitempty"`
	Users       []UserInfo      `json:"users,omitempty"`
	Docker      *DockerInfo     `json:"docker,omitempty"`
	CollectedAt string          `json:"collected_at"`
}

type HostInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	PlatformFamily  string `json:"platform_family"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
	Uptime          uint64 `json:"uptime_seconds"`
	BootTime        uint64 `json:"boot_time"`
	Procs           uint64 `json:"num_processes"`
}

type CPUInfo struct {
	Model      string    `json:"model"`
	Cores      int       `json:"physical_cores"`
	Threads    int       `json:"logical_cores"`
	UsagePerCPU []float64 `json:"usage_per_cpu_percent"`
	UsageTotal float64   `json:"usage_total_percent"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total_bytes"`
	TotalHuman  string  `json:"total"`
	Used        uint64  `json:"used_bytes"`
	UsedHuman   string  `json:"used"`
	Available   uint64  `json:"available_bytes"`
	AvailHuman  string  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
}

type SwapInfo struct {
	Total       uint64  `json:"total_bytes"`
	TotalHuman  string  `json:"total"`
	Used        uint64  `json:"used_bytes"`
	UsedHuman   string  `json:"used"`
	Free        uint64  `json:"free_bytes"`
	FreeHuman   string  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total_bytes"`
	TotalHuman  string  `json:"total"`
	Used        uint64  `json:"used_bytes"`
	UsedHuman   string  `json:"used"`
	Free        uint64  `json:"free_bytes"`
	FreeHuman   string  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkInfo struct {
	Interfaces []NetInterface `json:"interfaces"`
	IOCounters []NetIO        `json:"io_counters"`
}

type NetInterface struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"mac_address"`
	Addresses    []string `json:"addresses"`
	Flags        []string `json:"flags"`
	MTU          int      `json:"mtu"`
}

type NetIO struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	Errin     uint64 `json:"errors_in"`
	Errout    uint64 `json:"errors_out"`
}

type LoadInfo struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	Username   string  `json:"username"`
	CreateTime int64   `json:"create_time"`
	Cmdline    string  `json:"cmdline"`
}

type UserInfo struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  int    `json:"started"`
}

type DockerInfo struct {
	ServerVersion string          `json:"server_version"`
	Containers    []ContainerInfo `json:"containers"`
}

type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Ports   []string          `json:"ports"`
	Labels  map[string]string `json:"labels"`
	Created int64             `json:"created"`
}
