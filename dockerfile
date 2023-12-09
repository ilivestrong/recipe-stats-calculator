# Use the official Golang image to create a build artifact.
FROM golang:1.21.4-alpine3.18 as build-stage

# Set the Current Working Directory inside the container.
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container.
COPY . .

# Load dependencies
RUN go mod download && go mod verify

# Build the Go app.
RUN go build -o myapp .

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v -count=1 ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

# Copy the Pre-built binary file from the previous stage.
COPY --from=build-stage /app /app

# Set the Current Working Directory inside the container.
WORKDIR /app

# Command to run the executable.
ENTRYPOINT ["/app/myapp"]
