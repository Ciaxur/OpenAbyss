FROM golang:1.17

# Copy required app files
WORKDIR /go/src/app
COPY . .

# Build Server Binary
RUN go mod tidy
RUN mkdir build || \
  go build -o build/server ./server

# Run Built Binary
CMD ["./build/server"]
