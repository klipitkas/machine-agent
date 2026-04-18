package collector

import (
	"context"
	"testing"
)

func TestCollectDefaultSections(t *testing.T) {
	info := Collect(context.Background(), DefaultSections)
	if info.CollectedAt == "" {
		t.Error("CollectedAt should not be empty")
	}
	if info.Host == nil {
		t.Error("Host should be populated in default sections")
	}
	if info.Host != nil && info.Host.Hostname == "" {
		t.Error("Hostname should not be empty")
	}
	if info.CPU == nil {
		t.Error("CPU should be populated in default sections")
	}
	if info.Memory == nil {
		t.Error("Memory should be populated in default sections")
	}
	if info.Load == nil {
		t.Error("Load should be populated in default sections")
	}
	if info.Network == nil {
		t.Error("Network should be populated in default sections")
	}
}

func TestCollectSpecificSection(t *testing.T) {
	sections := map[string]bool{"host": true}
	info := Collect(context.Background(), sections)
	if info.Host == nil {
		t.Error("Host should be populated")
	}
	if info.CPU != nil {
		t.Error("CPU should not be populated when not requested")
	}
	if info.Memory != nil {
		t.Error("Memory should not be populated when not requested")
	}
}

func TestCollectEmptySections(t *testing.T) {
	sections := map[string]bool{}
	info := Collect(context.Background(), sections)
	if info.Host != nil {
		t.Error("Host should not be populated with empty sections")
	}
	if info.CollectedAt == "" {
		t.Error("CollectedAt should always be set")
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
		{17179869184, "16.0 GB"},
	}
	for _, tt := range tests {
		result := humanBytes(tt.input)
		if result != tt.expected {
			t.Errorf("humanBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCollectHost(t *testing.T) {
	info := &MachineInfo{}
	collectHost(context.Background(), info)

	if info.Host == nil {
		t.Fatal("Host should not be nil")
	}
	if info.Host.OS == "" {
		t.Error("OS should not be empty")
	}
	if info.Host.KernelArch == "" {
		t.Error("KernelArch should not be empty")
	}
}

func TestCollectCPU(t *testing.T) {
	info := &MachineInfo{}
	collectCPU(context.Background(), info)

	if info.CPU == nil {
		t.Fatal("CPU should not be nil")
	}
	if info.CPU.Threads <= 0 {
		t.Error("logical cores should be > 0")
	}
}

func TestCollectMemory(t *testing.T) {
	info := &MachineInfo{}
	collectMemory(context.Background(), info)

	if info.Memory == nil {
		t.Fatal("Memory should not be nil")
	}
	if info.Memory.Total == 0 {
		t.Error("total memory should be > 0")
	}
	if info.Memory.TotalHuman == "" {
		t.Error("human-readable total should not be empty")
	}
}
