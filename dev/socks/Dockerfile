FROM alpine:3.20 AS build
RUN apk add --no-cache git build-base \
 && git clone --depth 1 https://github.com/rofl0r/microsocks /src \
 && make -C /src

FROM alpine:3.20
COPY --from=build /src/microsocks /usr/local/bin/microsocks
USER 65534:65534
ENTRYPOINT ["microsocks"]
