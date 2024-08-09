FROM golang:1.22-alpine as builder
WORKDIR /
COPY . ./
RUN go mod download

RUN go build -o /commoner

FROM alpine
COPY --from=builder /commoner .

EXPOSE 80
CMD [ "/commoner" ]