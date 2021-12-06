#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.17.3-bullseye as builder
WORKDIR /go/src/app
COPY . .
RUN go build -o /go/bin/actions-runner-manager ./pkg

# Distroless: https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/base-debian11
MAINTAINER "Brett Logan"
LABEL org.opencontainers.image.source="https://github.com/lindluni/actions-runner-manager"
COPY --from=builder /go/bin/actions-runner-manager /
CMD ["/actions-runner-manager"]