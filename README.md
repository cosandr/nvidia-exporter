# NVIDIA Prometheus Exporter

This exporter used the NVidia Management Library (NVML) to query information
about the installed Nvidia GPUs.

Originally made by [BugRoger](https://github.com/BugRoger/nvidia-exporter).

Go module and some tweaks by [ashleyprimo](https://github.com/ashleyprimo/nvidia-exporter).

Additions with this fork:
* Export current graphics (`nvidia_clock_current_graphics`) and memory clock (`nvidia_clock_appdefault_graphics`)
* Export per-process utilization stats (pid, name, sm, mem, encoder, decoder), enable with `nvidia.per-process` option
* Export PCIe throughput `nvidia_pcie_tx_bytes` and `nvidia_pcie_rx_bytes`

## Requirements

The NVML shared library (libnvidia-ml.so.1) need to be loadable. When running
in a container it must be either baked in or mounted from the host.

## Running in Kubernetes

See [kubernetes.yaml](./kubernetes.yaml)

## Example

```
# HELP nvidia_clock_current_graphics Current GPU graphics clock speed as reported by the device
# TYPE nvidia_clock_current_graphics gauge
nvidia_clock_current_graphics{minor="0"} 570
# HELP nvidia_clock_current_memory Current GPU memory clock speed as reported by the device
# TYPE nvidia_clock_current_memory gauge
nvidia_clock_current_memory{minor="0"} 405
# HELP nvidia_device_count Count of found nvidia devices
# TYPE nvidia_device_count gauge
nvidia_device_count 1
# HELP nvidia_driver_info NVML Info
# TYPE nvidia_driver_info gauge
nvidia_driver_info{version="570.133.07"} 1
# HELP nvidia_fanspeed Fan speed as reported by the device
# TYPE nvidia_fanspeed gauge
nvidia_fanspeed{minor="0"} 0
# HELP nvidia_info Info as reported by the device
# TYPE nvidia_info gauge
nvidia_info{index="0",minor="0",name="GPU-27fb7f88-1ff1-d596-965b-3bc721e8b16d",uuid="NVIDIA GeForce RTX 4090"} 1
# HELP nvidia_memory_total Total memory as reported by the device
# TYPE nvidia_memory_total gauge
nvidia_memory_total{minor="0"} 2.5757220864e+10
# HELP nvidia_memory_used Used memory as reported by the device
# TYPE nvidia_memory_used gauge
nvidia_memory_used{minor="0"} 2.162622464e+09
# HELP nvidia_pcie_rx_bytes PCIe RX throughput as reported by the device
# TYPE nvidia_pcie_rx_bytes gauge
nvidia_pcie_rx_bytes{minor="0"} 850
# HELP nvidia_pcie_tx_bytes PCIe TX throughput as reported by the device
# TYPE nvidia_pcie_tx_bytes gauge
nvidia_pcie_tx_bytes{minor="0"} 1150
# HELP nvidia_power_limit Power limit as reported by the device in mW
# TYPE nvidia_power_limit gauge
nvidia_power_limit{minor="0"} 200000
# HELP nvidia_power_usage Power usage as reported by the device
# TYPE nvidia_power_usage gauge
nvidia_power_usage{minor="0"} 31037
# HELP nvidia_temperatures Temperature as reported by the device
# TYPE nvidia_temperatures gauge
nvidia_temperatures{minor="0"} 51
# HELP nvidia_up NVML Metric Collection Operational
# TYPE nvidia_up gauge
nvidia_up 1
# HELP nvidia_utilization_gpu GPU utilization as reported by the device
# TYPE nvidia_utilization_gpu gauge
nvidia_utilization_gpu{minor="0"} 23
# HELP nvidia_utilization_memory Memory Utilization as reported by the device
# TYPE nvidia_utilization_memory gauge
nvidia_utilization_memory{minor="0"} 27
# HELP nvidia_utilization_process_decutil Process decoder utilization stats averaged over 10s
# TYPE nvidia_utilization_process_decutil gauge
nvidia_utilization_process_decutil{minor="0",pid="2114"} 0
nvidia_utilization_process_decutil{minor="0",pid="3136"} 0
nvidia_utilization_process_decutil{minor="0",pid="845718"} 0
# HELP nvidia_utilization_process_encutil Process encoder utilization stats averaged over 10s
# TYPE nvidia_utilization_process_encutil gauge
nvidia_utilization_process_encutil{minor="0",pid="2114"} 0
nvidia_utilization_process_encutil{minor="0",pid="3136"} 0
nvidia_utilization_process_encutil{minor="0",pid="845718"} 0
# HELP nvidia_utilization_process_memutil Process memory utilization stats averaged over 10s
# TYPE nvidia_utilization_process_memutil gauge
nvidia_utilization_process_memutil{minor="0",pid="2114"} 17
nvidia_utilization_process_memutil{minor="0",pid="3136"} 1
nvidia_utilization_process_memutil{minor="0",pid="845718"} 16
# HELP nvidia_utilization_process_name Process name, if value is 0 the name couldn't be determined
# TYPE nvidia_utilization_process_name gauge
nvidia_utilization_process_name{minor="0",name="/opt/visual-studio-code/code",pid="845718"} 1
nvidia_utilization_process_name{minor="0",name="/usr/lib/Xorg",pid="2114"} 1
nvidia_utilization_process_name{minor="0",name="kitty",pid="3136"} 1
# HELP nvidia_utilization_process_smutil Process SM utilization stats averaged over 10s
# TYPE nvidia_utilization_process_smutil gauge
nvidia_utilization_process_smutil{minor="0",pid="2114"} 14
nvidia_utilization_process_smutil{minor="0",pid="3136"} 1
nvidia_utilization_process_smutil{minor="0",pid="845718"} 14
```
