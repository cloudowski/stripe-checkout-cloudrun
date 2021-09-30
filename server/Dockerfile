FROM golang:1.17-alpine AS build-env
ADD . /src
WORKDIR /src
RUN go build -o goapp

FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
EXPOSE 4242

ENV STATIC_DIR="sample-web" \
    DOMAIN="http://localhost:8080" \
    PAYMENT_METHOD_TYPES="p24,card" \
    STRIPE_COLLECT_ADDRESS="required" \
    STRIPE_COLLECT_TAXID="true" \
    STRIPE_ALLOW_PROMOCODES="true"

ENTRYPOINT [ "/app/goapp" ]