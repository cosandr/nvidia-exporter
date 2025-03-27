FROM nvcr.io/nvidia/cuda:12.4.1-base-ubuntu22.04

COPY nvidia-exporter /

ENTRYPOINT ["/nvidia-exporter"]
