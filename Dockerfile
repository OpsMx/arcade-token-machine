#
# Copyright 2021 OpsMx, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License")
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# Install the latest versions of our mods.  This is done as a separate step
# so it will pull from an image cache if possible, unless there are changes.
#
FROM --platform=${BUILDPLATFORM} golang:1.19-alpine AS buildmod
ENV CGO_ENABLED=0
RUN mkdir /build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

#
# Compile the code.
#
FROM buildmod AS build-binaries
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN mkdir /out
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /out/server app/server/*.go

#
# For a base image without an OS, this can be used:
#
FROM scratch AS base-image
COPY --from=alpine:3 /etc/ssl/cert.pem /etc/ssl/cert.pem

#
# Build the server image.  This should be a --target on docker build.
#
FROM base-image AS server-image
WORKDIR /app
COPY --from=build-binaries /out/server /app
ARG GIT_BRANCH
ENV GIT_BRANCH=${GIT_BRANCH}
ARG GIT_HASH
ENV GIT_HASH=${GIT_HASH}
EXPOSE 1982
CMD ["/app/server"]
