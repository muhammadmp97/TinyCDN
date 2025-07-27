# TinyCDN
<img width="2254" height="860" alt="tinycdn" src="https://github.com/user-attachments/assets/2c5884a7-a26e-4c5f-bec0-76773d712fff" />
TinyCDN is a simple Content Delivery Network (CDN) service implemented in Go. It supports caching with Redis, response compression with gzip, and includes monitoring integration with Prometheus.

## Features
- HTTP server with routing via [Gin](https://github.com/gin-gonic/gin)
- Caching layer backed by Redis for fast content delivery
- Gzip compression to reduce bandwidth usage
- Prometheus metrics for monitoring
- Fully containerized with Docker and Docker Compose

## Getting Started
```sh
git clone https://github.com/muhammadmp97/TinyCDN.git
cd TinyCDN
cp .env.example .env
docker-compose up --build
```

## Contributing
I started this project to get hands-on experience with Go and to learn how to cache and serve static files. I’d love to see contributions from others!
