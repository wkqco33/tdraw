# 낮은 glibc 베이스라인(기본 Ubuntu 20.04) 비디오 빌드용 이미지.
# 컨테이너 안에서 정적 FFmpeg와 libtcamviewer, tdraw를 모두 빌드해
# 결과 바이너리의 glibc 요구 버전을 낮춘다.
#
# 사용: docker build --build-arg GO_VERSION=1.27.1 -t tdraw-portable:20.04 -f docker/portable.Dockerfile docker
ARG UBUNTU_TAG=20.04
FROM ubuntu:${UBUNTU_TAG}

ARG GO_VERSION=1.27.1
ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake pkg-config nasm yasm \
        ca-certificates curl xz-utils git file \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz
ENV PATH=/usr/local/go/bin:/root/go/bin:/usr/local/bin:/usr/bin:/bin
ENV CGO_ENABLED=1 GOTOOLCHAIN=local

WORKDIR /w/tdraw
