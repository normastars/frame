FROM golang:1.17 as builder
WORKDIR /usr/src/app
ENV GOPROXY=https://goproxy.cn
# COPY go.mod ./
# COPY go.sum ./
# COPY manifest ./
# RUN go mod download
COPY . .
RUN make build

FROM scratch as server
WORKDIR /opt/app
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /usr/src/app/build/docker/Dockerfile /opt/app/Dockerfile
COPY --from=builder /usr/src/app/manifest.txt /opt/app/manifest.txt
COPY --from=builder /usr/src/app/conf/ /opt/app/conf/
COPY --from=builder /usr/src/app/bin/server /opt/app/
CMD ["/opt/app/server"]
