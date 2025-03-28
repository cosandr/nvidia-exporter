package main

import (
	"fmt"
	"strconv"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
)

type Metrics struct {
	Version string
	Devices []*Device
}

// Process contains the stats of a process running on the GPU
type Process struct {
	PID     uint32
	Name    *string
	SMUtil  uint32
	MemUtil uint32
	EncUtil uint32
	DecUtil uint32
}

// ToString returns a string representation of this process
func (p Process) ToString() string {
	var dbgStr = fmt.Sprintf("Process: %d, SM util: %d, Mem util: %d, Enc util: %d, Dec util: %d",
		p.PID, p.SMUtil, p.MemUtil, p.EncUtil, p.DecUtil)
	if p.Name != nil {
		dbgStr += fmt.Sprintf(", Name: %s", *p.Name)
	} else {
		dbgStr += ", Name: N/A"
	}
	return dbgStr
}

// PromPID returns the PID as a string for Prometheus
func (p Process) PromPID() string {
	return strconv.Itoa(int(p.PID))
}

type Device struct {
	Index                string
	MinorNumber          string
	Name                 string
	UUID                 string
	Temperature          float64
	PowerUsage           float64
	PowerLimit           float64
	FanSpeed             float64
	MemoryTotal          float64
	MemoryUsed           float64
	UtilizationMemory    float64
	UtilizationGPU       float64
	ClockCurrentGraphics float64
	ClockCurrentMemory   float64
	UtilizationProcesses []*Process
	PcieTxBytes          float64
	PcieRxBytes          float64
}

func collectMetrics() (*Metrics, error) {
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return nil, fmt.Errorf("failed to initialize nvml: %v", ret)
	}
	defer nvml.Shutdown()

	version, ret := nvml.SystemGetDriverVersion()
	if ret != nvml.SUCCESS {
		log.Warnf("Failed to get driver version: %v", ret)
	}

	metrics := &Metrics{
		Version: version,
	}

	numDevices, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("failed to get device count: %v", ret)
	}

	for index := range int(numDevices) {
		device, ret := nvml.DeviceGetHandleByIndex(index)
		if ret != nvml.SUCCESS {
			log.Errorf("failed to get device handle for GPU %d: %v", index, ret)
			continue
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			log.Errorf("failed to get device UUID for GPU %d: %v", index, ret)
			continue
		}

		name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			log.Errorf("failed to get device name for GPU %d: %v", index, ret)
			continue
		}

		minorNumber, ret := device.GetMinorNumber()
		if ret != nvml.SUCCESS {
			log.Errorf("failed to get device minor number for GPU %d: %v", index, ret)
			continue
		}

		temperature, temperatureErr := device.GetTemperature(nvml.TEMPERATURE_GPU)

		powerUsage, powerUsageErr := device.GetPowerUsage()

		powerLimit, powerLimitErr := device.GetEnforcedPowerLimit()

		fanSpeed, fanSpeedErr := device.GetFanSpeed()

		memoryInfo, memoryInfoErr := device.GetMemoryInfo()

		utilizationRates, utilizationRatesErr := device.GetUtilizationRates()

		clockCurrentGraphics, clockCurrentGraphicsErr := device.GetClock(nvml.CLOCK_GRAPHICS, nvml.CLOCK_ID_CURRENT)

		clockCurrentMemory, clockCurrentMemoryErr := device.GetClock(nvml.CLOCK_MEM, nvml.CLOCK_ID_CURRENT)

		pcieTxBytes, pcieTxBytesErr := device.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES)

		pcieRxBytes, pcieRxBytesErr := device.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES)

		var appendDevice = Device{
			Index:                strconv.Itoa(index),
			MinorNumber:          strconv.Itoa(int(minorNumber)),
			Name:                 name,
			UUID:                 uuid,
			Temperature:          checkError(temperatureErr, float64(temperature), index, "Temperature"),
			PowerUsage:           checkError(powerUsageErr, float64(powerUsage), index, "PowerUsage"),
			PowerLimit:           checkError(powerLimitErr, float64(powerLimit), index, "PowerLimit"),
			FanSpeed:             checkError(fanSpeedErr, float64(fanSpeed), index, "FanSpeed"),
			MemoryTotal:          checkError(memoryInfoErr, float64(memoryInfo.Total), index, "MemoryTotal"),
			MemoryUsed:           checkError(memoryInfoErr, float64(memoryInfo.Used), index, "MemoryUsed"),
			UtilizationMemory:    checkError(utilizationRatesErr, float64(utilizationRates.Memory), index, "UtilizationMemory"),
			UtilizationGPU:       checkError(utilizationRatesErr, float64(utilizationRates.Gpu), index, "UtilizationGPU"),
			ClockCurrentGraphics: checkError(clockCurrentGraphicsErr, float64(clockCurrentGraphics), index, "ClockCurrentGraphics"),
			ClockCurrentMemory:   checkError(clockCurrentMemoryErr, float64(clockCurrentMemory), index, "ClockCurrentMemory"),
			PcieTxBytes:          checkError(pcieTxBytesErr, float64(pcieTxBytes), index, "PcieTxBytes"),
			PcieRxBytes:          checkError(pcieRxBytesErr, float64(pcieRxBytes), index, "PcieRxBytes"),
		}
		// Skip process stats if not requested
		if !usePerProcess {
			metrics.Devices = append(metrics.Devices, &appendDevice)
			continue
		}
		// Collect per-process stats
		utilizations, utilizationsErr := device.GetProcessUtilization(10)
		if utilizationsErr != nvml.SUCCESS {
			log.Errorf("\tfailed to get process utilization for GPU %d: %v", index, utilizationsErr)
		} else {
			log.Debugf("process count: %d", len(utilizations))
			var pList []*Process
			for _, sample := range utilizations {
				var p = Process{
					PID:     sample.Pid,
					SMUtil:  sample.SmUtil,
					MemUtil: sample.MemUtil,
					EncUtil: sample.EncUtil,
					DecUtil: sample.DecUtil,
				}

				name, err := nvml.SystemGetProcessName(int(sample.Pid))
				if err != nvml.SUCCESS {
					log.Debugf("\tfailed to get process name for PID %d: %v\n", sample.Pid, err)
				} else {
					p.Name = &name
				}
				pList = append(pList, &p)
				log.Debug(p.ToString())
			}
			appendDevice.UtilizationProcesses = pList
		}
		metrics.Devices = append(metrics.Devices, &appendDevice)
	}
	return metrics, nil
}

// This function is used to check if error is returned
// if so set float64 to -1
func checkError(ret nvml.Return, value float64, index int, metric string) float64 {
	if ret != nvml.SUCCESS {
		log.Debugf("Unable to collect metrics for %s for device %d: %s", metric, index, ret)
		return -1
	}
	return value
}
