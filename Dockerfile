FROM ubuntu:24.04

COPY nvidia-exporter /

ENTRYPOINT ["/nvidia-exporter"]
