# TinyCDN
![tinycdn](https://github.com/user-attachments/assets/0c709d69-afbe-4386-a936-8f999254e7dd)
TinyCDN is a simple Content Delivery Network (CDN) service implemented in Go. It supports caching with Redis, response compression with gzip, and includes monitoring integration with Prometheus.

## Features
- HTTP server with routing via [Gin](https://github.com/gin-gonic/gin)
- Large files are stored externally on MinIO object storage
- Caching layer backed by Redis for fast content delivery
- Gzip compression to reduce bandwidth usage
- Prometheus metrics for monitoring
- Fully containerized with Docker and Docker Compose

## Getting Started
```sh
git clone https://github.com/muhammadmp97/TinyCDN.git
cd TinyCDN
cp .env.example .env
./tinycdn.sh up
curl -v http://localhost:8080/g/code.jquery.com?file=jquery-migrate-3.5.2.min.js
curl -v http://localhost:8080/g/images.pexels.com?file=photos/33060985/pexels-photo-33060985.jpeg
```

## Contributing
I started this project to get hands-on experience with Go and to learn how to cache and serve static files. I’d love to see contributions from others!
