FROM golang:1.21

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules and Sum Files
COPY go.mod ./
COPY go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o monitor .

# Install cron
RUN apt-get update && apt-get install -y cron

# Add crontab file in the cron directory
COPY crontab /etc/cron.d/monitor-cron

# Give execution rights on the cron job
RUN chmod 0644 /etc/cron.d/monitor-cron

# Apply cron job
RUN crontab /etc/cron.d/monitor-cron

# Create the log file to be able to run tail
RUN touch /var/log/cron.log

# Run the command on container startup
CMD cron && tail -f /var/log/cron.log