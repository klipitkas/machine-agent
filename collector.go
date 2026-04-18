package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

var defaultSections = map[string]bool{
	"host":   true,
	"cpu":    true,
	"memory": true,
	"load":   true,
	"network": true,
}

var allSections = map[string]func(context.Context, *MachineInfo){
	"host":      collectHost,
	"cpu":       collectCPU,
	"memory":    collectMemory,
	"swap":      collectSwap,
	"disks":     collectDisks,
	"network":   collectNetwork,
	"load":      collectLoad,
	"processes": collectProcesses,
	"users":     collectUsers,
	"docker":    collectDocker,
}

func collect(ctx context.Context, sections map[string]bool) (*MachineInfo, error) {
	info := &MachineInfo{
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
	}

	var collectors []func(context.Context, *MachineInfo)
	for name, fn := range allSections {
		if sections[name] {
			collectors = append(collectors, fn)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(collectors))
	for _, fn := range collectors {
		go func() {
			defer wg.Done()
			fn(ctx, info)
		}()
	}
	wg.Wait()

	return info, nil
}

func collectHost(ctx context.Context, info *MachineInfo) {
	h, err := host.InfoWithContext(ctx)
	if err != nil {
		return
	}
	info.Host = &HostInfo{
		Hostname:        h.Hostname,
		OS:              h.OS,
		Platform:        h.Platform,
		PlatformVersion: h.PlatformVersion,
		PlatformFamily:  h.PlatformFamily,
		KernelVersion:   h.KernelVersion,
		KernelArch:      h.KernelArch,
		Uptime:          h.Uptime,
		BootTime:        h.BootTime,
		Procs:           h.Procs,
	}
}

func collectCPU(ctx context.Context, info *MachineInfo) {
	c := &CPUInfo{}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err == nil && len(cpuInfo) > 0 {
		c.Model = cpuInfo[0].ModelName
	}

	physical, err := cpu.CountsWithContext(ctx, false)
	if err == nil {
		c.Cores = physical
	}

	logical, err := cpu.CountsWithContext(ctx, true)
	if err == nil {
		c.Threads = logical
	}

	percents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, true)
	if err == nil && len(percents) > 0 {
		c.UsagePerCPU = percents
		var sum float64
		for _, p := range percents {
			sum += p
		}
		c.UsageTotal = sum / float64(len(percents))
	}

	info.CPU = c
}

func collectMemory(ctx context.Context, info *MachineInfo) {
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		info.Memory = &MemoryInfo{
			Total:       v.Total,
			TotalHuman:  humanBytes(v.Total),
			Used:        v.Used,
			UsedHuman:   humanBytes(v.Used),
			Available:   v.Available,
			AvailHuman:  humanBytes(v.Available),
			UsedPercent: v.UsedPercent,
		}
	}
}

func collectSwap(ctx context.Context, info *MachineInfo) {
	s, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		info.Swap = &SwapInfo{
			Total:       s.Total,
			TotalHuman:  humanBytes(s.Total),
			Used:        s.Used,
			UsedHuman:   humanBytes(s.Used),
			Free:        s.Free,
			FreeHuman:   humanBytes(s.Free),
			UsedPercent: s.UsedPercent,
		}
	}
}

func collectDisks(ctx context.Context, info *MachineInfo) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return
	}

	for _, p := range partitions {
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}
		info.Disks = append(info.Disks, DiskInfo{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			Total:       usage.Total,
			TotalHuman:  humanBytes(usage.Total),
			Used:        usage.Used,
			UsedHuman:   humanBytes(usage.Used),
			Free:        usage.Free,
			FreeHuman:   humanBytes(usage.Free),
			UsedPercent: usage.UsedPercent,
		})
	}
}

func collectNetwork(ctx context.Context, info *MachineInfo) {
	n := &NetworkInfo{}

	ifaces, err := net.InterfacesWithContext(ctx)
	if err == nil {
		for _, iface := range ifaces {
			ni := NetInterface{
				Name:         iface.Name,
				HardwareAddr: iface.HardwareAddr,
				Flags:        iface.Flags,
				MTU:          iface.MTU,
			}
			for _, addr := range iface.Addrs {
				ni.Addresses = append(ni.Addresses, addr.Addr)
			}
			n.Interfaces = append(n.Interfaces, ni)
		}
	}

	counters, err := net.IOCountersWithContext(ctx, true)
	if err == nil {
		for _, c := range counters {
			n.IOCounters = append(n.IOCounters, NetIO{
				Name:        c.Name,
				BytesSent:   c.BytesSent,
				BytesRecv:   c.BytesRecv,
				PacketsSent: c.PacketsSent,
				PacketsRecv: c.PacketsRecv,
				Errin:       c.Errin,
				Errout:      c.Errout,
			})
		}
	}

	info.Network = n
}

func collectLoad(ctx context.Context, info *MachineInfo) {
	l, err := load.AvgWithContext(ctx)
	if err == nil {
		info.Load = &LoadInfo{
			Load1:  l.Load1,
			Load5:  l.Load5,
			Load15: l.Load15,
		}
	}
}

func collectProcesses(ctx context.Context, info *MachineInfo) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return
	}

	type procWithMem struct {
		proc   *process.Process
		memPct float32
	}

	candidates := make([]procWithMem, 0, len(procs))
	for _, p := range procs {
		memPct, err := p.MemoryPercentWithContext(ctx)
		if err != nil {
			continue
		}
		candidates = append(candidates, procWithMem{proc: p, memPct: memPct})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].memPct > candidates[j].memPct
	})
	if len(candidates) > 50 {
		candidates = candidates[:50]
	}

	procList := make([]ProcessInfo, len(candidates))
	var wg sync.WaitGroup
	wg.Add(len(candidates))
	for i, c := range candidates {
		go func() {
			defer wg.Done()
			p := c.proc
			name, _ := p.NameWithContext(ctx)
			status, _ := p.StatusWithContext(ctx)
			cpuPct, _ := p.CPUPercentWithContext(ctx)
			user, _ := p.UsernameWithContext(ctx)
			createTime, _ := p.CreateTimeWithContext(ctx)
			cmdline, _ := p.CmdlineWithContext(ctx)

			statusStr := ""
			if len(status) > 0 {
				statusStr = status[0]
			}

			procList[i] = ProcessInfo{
				PID:        p.Pid,
				Name:       name,
				Status:     statusStr,
				CPUPercent: cpuPct,
				MemPercent: c.memPct,
				Username:   user,
				CreateTime: createTime,
				Cmdline:    cmdline,
			}
		}()
	}
	wg.Wait()

	sort.Slice(procList, func(i, j int) bool {
		return procList[i].CPUPercent > procList[j].CPUPercent
	})
	info.Processes = procList
}

func collectUsers(ctx context.Context, info *MachineInfo) {
	users, err := host.UsersWithContext(ctx)
	if err != nil {
		return
	}
	for _, u := range users {
		info.Users = append(info.Users, UserInfo{
			User:     u.User,
			Terminal: u.Terminal,
			Host:     u.Host,
			Started:  u.Started,
		})
	}
}

func collectDocker(ctx context.Context, info *MachineInfo) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return
	}
	defer cli.Close()

	ping, err := cli.Ping(ctx)
	if err != nil {
		return
	}

	dockerInfo := &DockerInfo{
		ServerVersion: ping.APIVersion,
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		info.Docker = dockerInfo
		return
	}

	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		var ports []string
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		dockerInfo.Containers = append(dockerInfo.Containers, ContainerInfo{
			ID:      c.ID[:12],
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Ports:   ports,
			Labels:  c.Labels,
			Created: c.Created,
		})
	}

	info.Docker = dockerInfo
}
