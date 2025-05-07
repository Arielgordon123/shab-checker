# Start from the official Go image
FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-app

FROM alpine:latest  
 
RUN apk --no-cache add ca-certificates  logrotate

WORKDIR /root/

COPY logrotate.conf /etc/logrotate.conf

COPY --from=build /go-app  .
# Copy the run script
COPY run.sh /root/run.sh
RUN chmod +x /root/run.sh

# Run the script
CMD ["/bin/sh", "/root/run.sh"]