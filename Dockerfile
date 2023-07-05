FROM --platform=x86_64 golang:alpine as build

WORKDIR /usr/src/app
ADD * .
ENV CGO_ENABLED=0
RUN go build -v -o /rekor_crawler

FROM cgr.dev/chainguard/static:latest
COPY --from=build /rekor_crawler /
CMD ["/rekor_crawler"]
