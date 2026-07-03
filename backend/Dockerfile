FROM golang:1.25.4

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
  go build -o api ./cmd/api

EXPOSE 9000

CMD ["./api"]