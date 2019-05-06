# Start tho Go app build
FROM golang:latest AS build

# Copy source
WORKDIR /go/src/CSC482
COPY . .

# Get required modules (assumes packages have been added to ./vendor
RUN go get -d -v ./...

# Build a statcally-linked Go binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

# New build phase -- create binary-only image
FROM alpine:latest

# Add support for HTTPS and time zones
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates

WORKDIR /root/

# Copy files from prvious build container
COPY --from=build /go/src/CSC482/main ./

# Add environment variables
ENV LOGGLY_TOKEN 283eda33-38d2-4e6b-b161-d7ddf7731306
ENV AWS_ACCESS_KEY_ID AKIA34XNLPJYJ2SSHJUN
ENV AWS_SECRET_ACCESS_KEY 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

# Check results
RUN env && pwd && find .

# Start the application
CMD ["./main"]
