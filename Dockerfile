# Use an official Go image as the base
FROM --platform=$BUILDPLATFORM golang:alpine
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG GIT_COMMIT
ARG GIT_BRANCH
ARG IMAGE_NAME
ENV GIT_COMMIT=$GIT_COMMIT


# Set the working directory to /app
WORKDIR /app

# Copy the Go mod files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the application code
COPY . .
COPY entrypoint.sh .

# Build the application and inject the git commit
RUN go build  -ldflags "-X main.Commit=${GIT_COMMIT} -X main.Branch=${GIT_BRANCH}" -o comp2unraid

RUN chmod +x ./entrypoint.sh

# Create non-root user, output directory, and set ownership
RUN adduser -D -g '' app && mkdir /output && chown -R app:app /app /output
USER app
WORKDIR /output

ENTRYPOINT [ "/app/entrypoint.sh" ]

# Run the command to start the application
CMD ["--help"]
