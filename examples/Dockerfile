# Example multi-stage build: compiles dlib from source (Alpine has no dlib
# package), builds the detection example against it, and produces a slim
# runtime image with only the shared libraries the compiled binary needs.
#
# Build:
#   docker build -t go-recognizer-example -f examples/Dockerfile .
# Run (mount your own fotos/models directories):
#   docker run --rm \
#     -v "$(pwd)/examples/fotos":/data/fotos:ro \
#     -v "$(pwd)/examples/models":/data/models:ro \
#     -v "$(pwd)/out":/data \
#     go-recognizer-example

# ---- dlib: compile dlib from source ----
FROM golang:1.25-alpine AS dlib

RUN apk add --no-cache \
    cmake make g++ libstdc++ libgcc \
    jpeg-dev libpng-dev giflib-dev libjpeg-turbo-dev \
    blas-dev lapack-dev \
    ca-certificates wget

WORKDIR /
ARG DLIB_VERSION=v19.24.8

RUN wget -q https://github.com/davisking/dlib/archive/${DLIB_VERSION}.tar.gz \
 && tar xf ${DLIB_VERSION}.tar.gz \
 && mv dlib-* dlib \
 && mkdir -p dlib/build \
 # GCC 15+ dropped some implicit <cstdint> includes dlib relies on.
 && sed -i '1i#include <cstdint>' dlib/dlib/serialize.h \
 # CMake 4.x refuses dlib's old cmake_minimum_required() declarations.
 && find dlib -type f \( -name "CMakeLists.txt" -o -name "*.cmake" \) \
      -exec sed -i -E 's/cmake_minimum_required\s*\(\s*VERSION\s+[0-9.]+\s*\)/cmake_minimum_required(VERSION 3.10)/gI' {} + \
 && cd dlib/build \
 && cmake -DCMAKE_POLICY_VERSION_MINIMUM=3.10 \
          -DCMAKE_BUILD_TYPE=Release \
          -DDLIB_PNG_SUPPORT=ON \
          -DDLIB_GIF_SUPPORT=ON \
          -DDLIB_JPEG_SUPPORT=ON .. \
 && cmake --build . --config Release -- -j"$(nproc)" \
 && make install \
 && cd / && rm -rf dlib ${DLIB_VERSION}.tar.gz

# ---- builder: compile the Go example against dlib ----
FROM dlib AS builder

WORKDIR /src
COPY examples/ ./examples/

WORKDIR /src/examples
RUN go mod tidy \
 && CGO_ENABLED=1 go build -o /out/detection ./detection

# ---- runtime: slim image with just the compiled binary ----
FROM alpine:latest

# blas/cblas/lapack, libjpeg-turbo/libpng/giflib: dlib's runtime
# dependencies. dlib itself is statically linked into the binary (the
# cmake build above produces libdlib.a, not a .so), so it isn't needed
# here.
RUN apk add --no-cache \
    libstdc++ libgomp \
    libjpeg-turbo libpng giflib \
    blas cblas lapack

COPY --from=builder /out/detection /usr/local/bin/detection

WORKDIR /data
ENTRYPOINT ["/usr/local/bin/detection"]
