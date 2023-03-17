# Inspired by https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
# Create appuser.
ENV USER=appuser
ENV UID=10001 
# See https://stackoverflow.com/a/55757473/12429735RUN 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"
COPY . .
# Fetch dependencies.
ENV GOPATH "" 
RUN go mod download
RUN go mod verify

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/wsevent-service

############################
# STEP 2 build a small image
############################
FROM scratch
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our static executable.
COPY --from=builder /go/bin/wsevent-service /go/bin/wsevent-service
# Use an unprivileged user.
USER appuser:appuser
EXPOSE 8080
# Run the binary.
ENTRYPOINT ["/go/bin/wsevent-service"]