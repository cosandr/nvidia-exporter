FROM nvcr.io/nvidia/cuda:12.5.1-base-ubuntu24.04

COPY nvidia-exporter /

ENTRYPOINT ["/nvidia-exporter"]
