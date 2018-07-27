# ---- Base Node ----
FROM golang:1.10 AS base

RUN git clone https://github.com/kubernetes/kubernetes.git
RUN apt-get update && apt-get install -y rsync
RUN cd kubernetes && \
    make all WHAT=cmd/kube-scheduler && \
    cp ./_output/bin/kube-scheduler /kube-scheduler-hawq

# ---- Scheduler ----
FROM ubuntu:latest
COPY --from=0 /kube-scheduler-hawq /
ADD config.yaml /opt
ADD policy.json /opt