FROM golang:latest AS builder
WORKDIR /app
ADD . /app/
RUN make alpine

FROM alpine:latest
ENV LISTEN_PORT=8080

WORKDIR /proxy

COPY --from=builder /app/output/oauth2-authenticating-proxy .
CMD [ "./oauth2-authenticating-proxy", "serve" ]
