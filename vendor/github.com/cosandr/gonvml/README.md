Go Bindings for NVML
--------------------

[NVML or NVIDIA Management
Library](https://developer.nvidia.com/nvidia-management-library-nvml) is a
C-based API that can be used for monitoring NVIDIA GPU devices. It's closed
source but can be downloaded as part of the [GPU Deployment
Kit](https://developer.nvidia.com/gpu-deployment-kit).

The [NVML API
Reference](http://docs.nvidia.com/deploy/nvml-api/nvml-api-reference.html)
describe various methods that are available as part of NVML.

The `nvml.h` file is included in this repository so that we don't depend on
the presence of NVML in the build environment.

The `bindings.go` file is the cgo bridge which calls the NVML functions. The
cgo preamble in `bindings.go` uses `dlopen` to dynamically load NVML and makes
its functions available.

This fork adds the ability to query clock speeds as well, on my 1080 Ti I can access the following:

* Current graphics clock (actual GPU clock speed)
* Application default graphics clock (this appears to be the max memory clock speed)
* Application target graphics clock (similar to actual GPU clock speed)
* Boost max graphics clock (base clock?)

All other values result in either invalid argument or not supported error, quite probable they only work on Quadro or similar.

I've also added PcieThroughput queries.
