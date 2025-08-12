FROM golang:alpine3.22 AS base

WORKDIR /mr_robot

COPY . .

RUN apk update -y --no-cache \
  && apk upgrade -y --no-cache \
  && apk add git --no-cache \
  && go mod download \
  && go build -o /usr/local/bin/mr_robot ./cmd/mr_robot/main.go

# Development stage
FROM base AS development

RUN go install github.com/air-verse/air@latest

EXPOSE 8888

ENTRYPOINT [ "/go/bin/air" ]
CMD [ "-c", "/mr_robot/build/air.toml" ]

# Production runtime stage
FROM alpine:3.22 AS production

COPY --from=base /usr/local/bin/mr_robot /usr/local/bin/mr_robot
COPY --from=base /mr_robot/config /config

RUN apk update --no-cache \
  && apk upgrade --no-cache \
  && apk add ca-certificates --no-cache \
  && update-ca-certificates \
  && rm -rf /var/cache/apk/* \
  && rm -rf /tmp/* \
  && chmod +x /usr/local/bin/mr_robot \
  && chown nobody:nobody /usr/local/bin/mr_robot

USER nobody:nobody

EXPOSE 8888

ENTRYPOINT ["/usr/local/bin/mr_robot"]
