FROM golang:1.13 as builder
COPY . /src/
WORKDIR /src/
RUN go build -o hugepageschecker


FROM alpine:3.9
COPY --from=builder /src/hugepageschecker /hugepageschecker
WORKDIR /
CMD /hugepageschecker
ENTRYPOINT /hugepageschecker

