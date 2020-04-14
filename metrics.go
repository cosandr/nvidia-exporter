package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosandr/gonvml"
	log "github.com/sirupsen/logrus"
)

var averageDuration = 10 * time.Second

type Metrics struct {
	Version string
	Devices []*Device
}

// Process contains the stats of a process running on the GPU
type Process struct {
	PID     uint
	Name    *string
	SMUtil  uint
	MemUtil uint
	EncUtil uint
	DecUtil uint
}

// ToString returns a string representation of this process
func (p Process) ToString() string {
	var dbgStr = fmt.Sprintf("Process: %d, SM  util: %d, Mem util: %d, Enc util: %d, Dec util: %d",
		p.PID, p.SMUtil, p.MemUtil, p.EncUtil, p.DecUtil)
	if p.Name != nil {
		dbgStr += fmt.Sprintf(", Name: %s", *p.Name)
	} else {
		dbgStr += fmt.Sprintf(", Name: N/A")
	}
	return dbgStr
}

// PromPID returns the PID as a string for Prometheus
func (p Process) PromPID() string {
	return strconv.Itoa(int(p.PID))
}

type Device struct {
	Index                   string
	MinorNumber             string
	Name                    string
	UUID                    string
	Temperature             float64
	PowerUsage              float64
	PowerUsageAverage       float64
	FanSpeed                float64
	MemoryTotal             float64
	MemoryUsed              float64
	UtilizationMemory       float64
	UtilizationGPU          float64
	UtilizationGPUAverage   float64
	ClockCurrentGraphics    float64
	ClockAppDefaultGraphics float64
	UtilizationProcesses    []*Process
	PcieTxBytes             float64
	PcieRxBytes             float64
}

func collectMetrics() (*Metrics, error) {
	if err := gonvml.Initialize(); err != nil {
		log.Errorf("Failed to initialize gonvml.")
		// Return out, since this failure to initialize
		// will prevent collection.
		return nil, err
	}
	defer gonvml.Shutdown()

	version, err := gonvml.SystemDriverVersion()
	if err != nil {
		log.Warnf("Failed to get SystemDriverVersion.")
	}

	metrics := &Metrics{
		Version: version,
	}

	numDevices, err := gonvml.DeviceCount()
	if err != nil {
		log.Errorf("Failed to get DeviceCount")
		// Return out, since this failure to obtain
		// device count will prevent collection.
		return nil, err
	}

	for index := 0; index < int(numDevices); index++ {
		device, err := gonvml.DeviceHandleByIndex(uint(index))
		if err != nil {
			log.Errorf("Failed to get DeviceHandleByIndex")
			// Return out, since this failure to obtain
			// DeviceHandleByIndex will prevent collection.
			return nil, err
		}

		uuid, err := device.UUID()
		if err != nil {
			log.Errorf("Failed to get deviceUUID")
			// Return out, since this failure to obtain
			// failure to get this metrics is likely.
			// is of a problem.
			return nil, err
		}

		name, err := device.Name()
		if err != nil {
			log.Errorf("Failed to get deviceName")
			// Return out, since this failure to obtain
			// failure to get this metrics is likely.
			// is of a problem.
			return nil, err
		}

		minorNumber, err := device.MinorNumber()
		if err != nil {
			log.Errorf("Failed to get MinorNumber")
			// Return out, since this failure to obtain
			// MinorNumber will potentially cause conlficts.
			return nil, err
		}

		temperature, temperatureErr := device.Temperature()

		powerUsage, powerUsageErr := device.PowerUsage()

		powerUsageAverage, powerUsageAverageErr := device.AveragePowerUsage(averageDuration)

		fanSpeed, fanSpeedErr := device.FanSpeed()

		memoryTotal, memoryUsed, memoryInfoErr := device.MemoryInfo()

		utilizationGPU, utilizationMemory, utilizationRatesErr := device.UtilizationRates()

		utilizationGPUAverage, utilizationGPUAverageErr := device.AverageGPUUtilization(averageDuration)

		clockCurrentGraphics, clockCurrentGraphicsErr := device.Clock(gonvml.ClockIDCurrent, gonvml.ClockTypeGraphics)

		// This appears to be the memory clock
		clockAppDefaultGraphics, clockAppDefaultGraphicsErr := device.Clock(gonvml.ClockIDAppClockDefault, gonvml.ClockTypeGraphics)

		pcieTxBytes, pcieTxBytesErr := device.PcieThroughput(gonvml.PcieUtilTxBytes)

		pcieRxBytes, pcieRxBytesErr := device.PcieThroughput(gonvml.PcieUtilRxBytes)

		var appendDevice = Device{
			Index:                   strconv.Itoa(index),
			MinorNumber:             strconv.Itoa(int(minorNumber)),
			Name:                    name,
			UUID:                    uuid,
			Temperature:             checkError(temperatureErr, float64(temperature), index, "Temperature"),
			PowerUsage:              checkError(powerUsageErr, float64(powerUsage), index, "PowerUsage"),
			PowerUsageAverage:       checkError(powerUsageAverageErr, float64(powerUsageAverage), index, "PowerUsageAverage"),
			FanSpeed:                checkError(fanSpeedErr, float64(fanSpeed), index, "FanSpeed"),
			MemoryTotal:             checkError(memoryInfoErr, float64(memoryTotal), index, "MemoryTotal"),
			MemoryUsed:              checkError(memoryInfoErr, float64(memoryUsed), index, "MemoryUsed"),
			UtilizationMemory:       checkError(utilizationRatesErr, float64(utilizationMemory), index, "UtilizationMemory"),
			UtilizationGPU:          checkError(utilizationRatesErr, float64(utilizationGPU), index, "UtilizationGPU"),
			UtilizationGPUAverage:   checkError(utilizationGPUAverageErr, float64(utilizationGPUAverage), index, "UtilizationGPUAverage"),
			ClockCurrentGraphics:    checkError(clockCurrentGraphicsErr, float64(clockCurrentGraphics), index, "ClockCurrentGraphics"),
			ClockAppDefaultGraphics: checkError(clockAppDefaultGraphicsErr, float64(clockAppDefaultGraphics), index, "ClockAppDefaultGraphics"),
			PcieTxBytes:             checkError(pcieTxBytesErr, float64(pcieTxBytes), index, "PcieTxBytes"),
			PcieRxBytes:             checkError(pcieRxBytesErr, float64(pcieRxBytes), index, "PcieRxBytes"),
		}
		// Skip process stats if not requested
		if !usePerProcess {
			metrics.Devices = append(metrics.Devices, &appendDevice)
			continue
		}
		// Collect per-process stats
		utilizations, utilizationsErr := device.ProcessUtilization(10, averageDuration)
		if utilizationsErr != nil {
			// appendDevice.UtilizationProcesses = nil
			log.Errorf("\tdev.UtilizationProcesses() error: %v\n", err)
		} else {
			log.Debugf("Process count: %v\n", len(utilizations))
			var pList []*Process
			for _, sample := range utilizations {
				var p = Process{
					PID:     sample.Pid,
					SMUtil:  sample.SMUtil,
					MemUtil: sample.MemUtil,
					EncUtil: sample.EncUtil,
					DecUtil: sample.DecUtil,
				}

				name, err := gonvml.SystemGetProcessName(sample.Pid, 64)
				if err != nil {
					log.Errorf("\tdev.SystemGetProcessName() error: %v\n", err)
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
func checkError(err error, value float64, index int, metric string) float64 {
	if err != nil {
		log.Debugf("Unable to collect metrics for %s for device %d: %s", metric, index, err)
		return -1
	}
	return value
}
