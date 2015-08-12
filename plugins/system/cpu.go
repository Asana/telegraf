package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins"
	"github.com/influxdb/telegraf/plugins/system/ps/cpu"
)

type CPUStats struct {
	ps        PS
	lastStats []cpu.CPUTimesStat

	PerCPU   bool `toml:"percpu"`
	TotalCPU bool `toml:"totalcpu"`
}

func NewCPUStats(ps PS) *CPUStats {
	return &CPUStats{
		ps: ps,
	}
}

func (_ *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

var sampleConfig = `
# Whether to report per-cpu stats or not
percpu = true
# Whether to report total system cpu stats or not
totalcpu = true`

func (_ *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (s *CPUStats) Gather(acc plugins.Accumulator) error {
	times, err := s.ps.CPUTimes(s.PerCPU, s.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}

	for i, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		total := totalCpuTime(cts)

		// Add total cpu numbers
		fields := map[string]interface{}{
			"user":      cts.User,
			"system":    cts.System,
			"idle":      cts.Idle,
			"nice":      cts.Nice,
			"iowait":    cts.Iowait,
			"irq":       cts.Irq,
			"softirq":   cts.Softirq,
			"steal":     cts.Steal,
			"guest":     cts.Guest,
			"guestNice": cts.GuestNice,
			"stolen":    cts.Stolen,
		}

		acc.AddValues("raw", fields, tags)

		// Add in percentage
		if len(s.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU stats yet
			continue
		}
		lastCts := s.lastStats[i]
		lastTotal := totalCpuTime(lastCts)
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			return fmt.Errorf("Error: current total CPU time is less than previous total CPU time")
		}

		if totalDelta == 0 {
			continue
		}

		percentageFields := map[string]interface{}{
			"user":      100 * (cts.User - lastCts.User) / totalDelta,
			"system":    100 * (cts.System - lastCts.System) / totalDelta,
			"idle":      100 * (cts.Idle - lastCts.Idle) / totalDelta,
			"nice":      100 * (cts.Nice - lastCts.Nice) / totalDelta,
			"iowait":    100 * (cts.Iowait - lastCts.Iowait) / totalDelta,
			"irq":       100 * (cts.Irq - lastCts.Irq) / totalDelta,
			"softirq":   100 * (cts.Softirq - lastCts.Softirq) / totalDelta,
			"steal":     100 * (cts.Steal - lastCts.Steal) / totalDelta,
			"guest":     100 * (cts.Guest - lastCts.Guest) / totalDelta,
			"guestNice": 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta,
			"stolen":    100 * (cts.Stolen - lastCts.Stolen) / totalDelta,
		}

		acc.AddValues("percentage", percentageFields, tags)

	}

	s.lastStats = times

	return nil
}

func totalCpuTime(t cpu.CPUTimesStat) float64 {
	return t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal +
		t.Guest + t.GuestNice + t.Stolen + t.Idle
}

func init() {
	plugins.Add("cpu", func() plugins.Plugin {
		return &CPUStats{ps: &systemPS{}}
	})
}
