# Contact Share

A simple web application to share your personal contact information via QR code and downloadable vCard.

## Contact Information Supported

- Email
- Phone Number
- Bluesky Profile
- GitHub Profile
- WhatsApp
- Facebook Profile

## Prerequisites

- Go 1.25 or higher
- A `.env` file with your personal information (see Setup below)

## Setup

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd contactshare
   ```

2. **Copy the example environment file**
   ```bash
   cp .env.example .env
   ```

3. **Edit `.env` with your personal information**
   ```bash
   # Required
   NAME="Your Name"
   
   # Optional - Add the ones you want to share
   EMAIL="your.email@example.com"
   PHONE="+1234567890"
   BLUESKY="https://bsky.app/profile/your.handle"
   GITHUB="https://github.com/yourusername"
   WHATSAPP="https://wa.me/1234567890"
   FACEBOOK="https://facebook.com/yourprofile"
   
   # Server configuration
   BASE_URL="https://info.base.com"
   PORT=8080
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

   The server will start on `http://localhost:8080`

## Building for Production

```bash
go build -o contactshare main.go
./contactshare
```

## Docker Deployment

1. **Build the Docker image**
   ```bash
   docker build -t contactshare .
   ```

2. **Run the container**
   ```bash
   docker run -d -p 8080:8080 --env-file .env contactshare
   ```

   Or with environment variables:
   ```bash
   docker run -d -p 8080:8080 \
     -e NAME="Your Name" \
     -e EMAIL="your@email.com" \
     -e BASE_URL="https://info.base.com" \
     contactshare
   ```

## Endpoints

- `GET /` - Main contact information page
- `GET /qr` - QR code image (PNG) pointing to your BASE_URL
- `GET /contact.vcf` - Download vCard file
- `GET /health` - Health check endpoint (returns JSON status)
- `GET /favicon.ico` - Favicon (1x1 transparent PNG)

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `NAME` | Yes | Your full name | "John Doe" |
| `EMAIL` | No | Email address | "john@example.com" |
| `PHONE` | No | Phone number | "+1234567890" |
| `BLUESKY` | No | Bluesky profile URL | "https://bsky.app/profile/user.bsky.social" |
| `GITHUB` | No | GitHub profile URL | "https://github.com/username" |
| `WHATSAPP` | No | WhatsApp chat URL | "https://wa.me/1234567890" |
| `FACEBOOK` | No | Facebook profile URL | "https://facebook.com/username" |
| `BASE_URL` | No | Your domain URL | "https://info.base.com" |
| `PORT` | No | Server port (default: 8080) | "8080" |

## Deployment

### Systemd Service (Linux)

Create `/etc/systemd/system/contactshare.service`:

```ini
[Unit]
Description=Contact Share Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/contactshare
EnvironmentFile=/opt/contactshare/.env
ExecStart=/opt/contactshare/contactshare
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable contactshare
sudo systemctl start contactshare
```

### Reverse Proxy (Caddy)

```caddy
info.base.com {
    reverse_proxy localhost:8080
}
```


## License

MIT
