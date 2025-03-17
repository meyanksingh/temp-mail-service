# Temp Mail Service

A lightweight, self-hosted temporary email service written in Go. This service provides disposable email addresses that can receive emails through SMTP and makes them accessible via a RESTful API.

## Features

- **SMTP Server**: Accepts incoming emails on port 25
- **HTTP API**: Provides RESTful endpoints to access received emails
- **Redis Storage**: Stores emails in Redis for fast access and automatic expiration
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **CORS Support**: Configurable CORS for frontend integration
- **Customizable Domain**: Use your own domain for temporary email addresses

## Architecture

The service consists of two main components:

1. **SMTP Server**: Listens for incoming emails, parses them, and stores them in Redis
2. **HTTP Server**: Provides API endpoints to access and manage temporary emails

## Prerequisites

- Go 1.23.5 or higher
- Redis server
- Docker and Docker Compose (for containerized deployment)
- A domain name with proper DNS configuration (for production use)

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| HOST | Domain name for the email service | localhost |
| PORT | SMTP server port | 25 |
| HTTP_PORT | HTTP API server port | 8000 |
| LOG_LEVEL | Logging level (debug, info, warn, error) | info |
| REDIS_URL | Redis connection URL | redis://localhost:6379 |
| ALLOWED_ORIGINS | CORS allowed origins (comma-separated) | * |

## Installation

### Using Docker Compose (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/meyanksingh/temp-mail-service.git
   cd temp-mail-service
   ```

2. Configure environment variables in `docker-compose.yml` or create a `.env` file.

3. Start the services:
   ```bash
   docker-compose up -d
   ```

### Manual Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/meyanksingh/temp-mail-service.git
   cd temp-mail-service
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file with your configuration.

4. Build and run:
   ```bash
   go build -o temp-mail-service
   ./temp-mail-service
   ```

## DNS Configuration

To use your own domain with this service, you need to configure DNS records properly:

### MX Record

Add an MX (Mail Exchange) record to your domain's DNS settings:

```
Type: MX
Host: @ (or your subdomain, e.g., mail)
Priority: 10
Value: your-server-hostname.com (or subdomain.your-domain.com)
TTL: 3600 (or as preferred)
```

### A Record

Add an A record pointing to your server's IP address:

```
Type: A
Host: @ (or your subdomain, e.g., mail)
Value: YOUR_SERVER_IP_ADDRESS
TTL: 3600 (or as preferred)
```

For example, if you want to use `mail.example.com` for your temporary email service:
1. Create an A record: `mail.example.com` → `YOUR_SERVER_IP_ADDRESS`
2. Create an MX record: `mail.example.com` with priority 10 pointing to `mail.example.com`

Note: DNS changes may take up to 24-48 hours to propagate globally.

## API Endpoints

### Generate a Random Email

```
GET /api/email
```

Generates a random email address on your configured domain.

**Response:**
```json
{
  "email": "random123@yourdomain.com"
}
```

### Get Emails for an Address

```
GET /api/email/:address
```

Retrieves all emails received for the specified address.

**Response:**
```json
[
  {
    "from": "sender@example.com",
    "to": "random123@yourdomain.com",
    "subject": "Test Email",
    "body": "This is a test email",
    "createdAt": "2023-03-13T12:34:56Z"
  }
]
```

## Docker Deployment

The included Dockerfile and docker-compose.yml make it easy to deploy the service:

```bash
# Build the Docker image
docker build -t temp-mail-service .

# Run with Docker
docker run -p 25:25 -p 8000:8000 \
  -e HOST=yourdomain.com \
  -e PORT=25 \
  -e HTTP_PORT=8000 \
  -e REDIS_URL=redis://redis-server:6379 \
  temp-mail-service
```

## Testing

### Using Telnet

You can test the SMTP server using telnet:

```bash
# Connect to the SMTP server
telnet yourdomain.com 25

# Once connected, you can send a test email with these commands:
HELO client.example.com
MAIL FROM:<sender@example.com>
RCPT TO:<test@yourdomain.com>
DATA
Subject: Test Email

This is a test email body.
.
QUIT
```

### Using Swaks (Swiss Army Knife for SMTP)

[Swaks](https://github.com/jetmore/swaks) is a versatile SMTP testing tool:

```bash
# Install swaks (on Debian/Ubuntu)
apt-get install swaks

# On macOS with Homebrew
brew install swaks

# Send a test email
swaks --server yourdomain.com \
      --port 25 \
      --from sender@example.com \
      --to test@yourdomain.com \
      --header "Subject: Test Email" \
      --body "This is a test email sent with swaks."
```

### Testing the API

You can test the HTTP API using curl:

```bash
# Generate a random email
curl http://yourdomain.com:8000/api/email

# Get emails for an address
curl http://yourdomain.com:8000/api/email/test@yourdomain.com
```

## Frontend Integration

You can integrate this service with any frontend application by configuring the CORS settings:

```
ALLOWED_ORIGINS=https://yourdomain.com,http://localhost:3000
```

## Security Considerations

- This service is designed for temporary email usage and should not be used for sensitive communications
- By default, the SMTP server accepts anonymous connections
- Consider running behind a reverse proxy with TLS for production use
- For production, consider using a firewall to restrict access to port 25 to prevent abuse

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [go-smtp](https://github.com/emersion/go-smtp) - SMTP server library
- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [go-redis](https://github.com/redis/go-redis) - Redis client for Go
