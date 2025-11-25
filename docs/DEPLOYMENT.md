# Deployment Guide

This guide covers deploying Rockstar Web Framework applications to production environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Building for Production](#building-for-production)
3. [Configuration](#configuration)
4. [Database Setup](#database-setup)
5. [Deployment Options](#deployment-options)
6. [Monitoring](#monitoring)
7. [Security Checklist](#security-checklist)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Go**: 1.25 or higher
- **Operating System**: Linux, macOS, Windows, AIX, Unix
- **Memory**: Minimum 512MB RAM (2GB+ recommended)
- **CPU**: 1+ cores (4+ recommended for production)
- **Disk**: 100MB+ for application, varies with database

### External Dependencies

- **Database**: MySQL 5.7+, PostgreSQL 10+, MSSQL 2017+, or SQLite 3+
- **Cache** (optional): Redis 5.0+
- **TLS Certificates**: For HTTPS/QUIC support

## Building for Production

### 1. Build the Application

```bash
# Standard build
go build -o app cmd/rockstar/main.go

# Optimized build with reduced binary size
go build -ldflags="-s -w" -o app cmd/rockstar/main.go

# Cross-compilation for Linux
GOOS=linux GOARCH=amd64 go build -o app-linux cmd/rockstar/main.go
```

### 2. Build with Version Information

```bash
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

go build -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
    -o app cmd/rockstar/main.go
```

### 3. Create Distribution Package

```bash
# Create directory structure
mkdir -p dist/app
mkdir -p dist/app/config
mkdir -p dist/app/locales
mkdir -p dist/app/logs

# Copy files
cp app dist/app/
cp config.yaml dist/app/config/
cp locales/*.yaml dist/app/locales/

# Create archive
tar -czf app-v1.0.0.tar.gz -C dist app
```

## Configuration

### Production Configuration File

Create `config.yaml`:

```yaml
server:
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 120s
  max_header_bytes: 2097152  # 2MB
  enable_http1: true
  enable_http2: true
  enable_quic: false
  shutdown_timeout: 30s

database:
  driver: postgres
  host: db.example.com
  port: 5432
  database: production_db
  username: app_user
  password: ${DB_PASSWORD}  # From environment
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

cache:
  type: redis
  host: cache.example.com
  port: 6379
  password: ${REDIS_PASSWORD}
  max_size: 524288000  # 500MB
  default_ttl: 10m

session:
  storage: database
  cookie_name: app_session
  expiration: 24h
  secure: true
  http_only: true
  same_site: strict

security:
  enable_xframe_options: true
  xframe_options: SAMEORIGIN
  enable_cors: true
  enable_csrf: true
  enable_xss: true
  max_request_size: 20971520  # 20MB
  request_timeout: 60s

monitoring:
  enable_metrics: true
  metrics_path: /metrics
  enable_pprof: false  # Disable in production
  pprof_path: /debug/pprof

logging:
  level: info
  format: json
  output: /var/log/app/app.log
```

### Environment Variables

Create `.env` file:

```bash
# Database
DB_PASSWORD=secure_password_here
DB_HOST=db.example.com

# Cache
REDIS_PASSWORD=redis_password_here

# Application
APP_ENV=production
APP_PORT=8080

# TLS
TLS_CERT=/etc/ssl/certs/app.crt
TLS_KEY=/etc/ssl/private/app.key
```

## Database Setup

### PostgreSQL

```sql
-- Create database
CREATE DATABASE production_db;

-- Create user
CREATE USER app_user WITH PASSWORD 'secure_password';

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE production_db TO app_user;

-- Create tables (run migrations)
\c production_db
\i migrations/001_initial_schema.sql
```

### MySQL

```sql
-- Create database
CREATE DATABASE production_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'app_user'@'%' IDENTIFIED BY 'secure_password';

-- Grant permissions
GRANT ALL PRIVILEGES ON production_db.* TO 'app_user'@'%';
FLUSH PRIVILEGES;

-- Run migrations
USE production_db;
SOURCE migrations/001_initial_schema.sql;
```

### Database Migrations

Create migration files in `migrations/` directory:

```sql
-- migrations/001_initial_schema.sql
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255),
    tenant_id VARCHAR(255),
    data TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Add more tables...
```

## Deployment Options

### Option 1: Systemd Service (Linux)

Create `/etc/systemd/system/rockstar-app.service`:

```ini
[Unit]
Description=Rockstar Web Application
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=app
Group=app
WorkingDirectory=/opt/app
Environment="APP_ENV=production"
EnvironmentFile=/opt/app/.env
ExecStart=/opt/app/app -addr :8080 -config /opt/app/config/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/app

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable rockstar-app
sudo systemctl start rockstar-app
sudo systemctl status rockstar-app
```

### Option 2: Docker Container

Create `Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app cmd/rockstar/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app .
COPY --from=builder /build/config ./config
COPY --from=builder /build/locales ./locales

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app && \
    chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["./app"]
CMD ["-addr", ":8080", "-config", "config/config.yaml"]
```

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    networks:
      - app-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=production_db
      - POSTGRES_USER=app_user
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    networks:
      - app-network

volumes:
  postgres-data:
  redis-data:

networks:
  app-network:
    driver: bridge
```

Deploy:

```bash
docker-compose up -d
docker-compose logs -f app
```

### Option 3: Kubernetes

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rockstar-app
  labels:
    app: rockstar-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rockstar-app
  template:
    metadata:
      labels:
        app: rockstar-app
    spec:
      containers:
      - name: app
        image: your-registry/rockstar-app:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: db-password
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: rockstar-app
spec:
  selector:
    app: rockstar-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

Deploy:

```bash
kubectl apply -f k8s/deployment.yaml
kubectl get pods
kubectl logs -f deployment/rockstar-app
```

### Option 4: Reverse Proxy (Nginx)

Create `/etc/nginx/sites-available/rockstar-app`:

```nginx
upstream rockstar_backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://rockstar_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location /metrics {
        deny all;  # Restrict metrics endpoint
    }
}
```

Enable and reload:

```bash
sudo ln -s /etc/nginx/sites-available/rockstar-app /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Monitoring

### Health Checks

The framework provides built-in health check endpoints:

```bash
# Liveness check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready
```

### Metrics

Access Prometheus-compatible metrics:

```bash
curl http://localhost:8080/metrics
```

### Logging

Configure structured logging:

```go
app.RegisterStartupHook(func(ctx context.Context) error {
    logger := app.Logger()
    logger.Info("Application started",
        "version", version,
        "environment", env,
    )
    return nil
})
```

### Profiling

Enable pprof for performance analysis (development only):

```bash
# CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Memory profile
curl http://localhost:8080/debug/pprof/heap > mem.prof

# Analyze
go tool pprof cpu.prof
```

## Security Checklist

### Pre-Deployment

- [ ] Update all dependencies
- [ ] Enable HTTPS/TLS
- [ ] Configure secure session cookies
- [ ] Enable CSRF protection
- [ ] Configure CORS properly
- [ ] Set secure headers (X-Frame-Options, etc.)
- [ ] Disable debug endpoints (pprof)
- [ ] Use strong database passwords
- [ ] Enable rate limiting
- [ ] Configure request size limits
- [ ] Set request timeouts
- [ ] Review and minimize permissions

### Post-Deployment

- [ ] Monitor logs for errors
- [ ] Check metrics for anomalies
- [ ] Verify health checks
- [ ] Test failover scenarios
- [ ] Review security headers
- [ ] Audit access logs
- [ ] Update firewall rules
- [ ] Configure backup strategy

## Troubleshooting

### Application Won't Start

```bash
# Check logs
journalctl -u rockstar-app -n 50

# Check configuration
./app -config config/config.yaml --validate

# Check port availability
netstat -tuln | grep 8080
```

### Database Connection Issues

```bash
# Test database connectivity
psql -h db.example.com -U app_user -d production_db

# Check connection pool
curl http://localhost:8080/metrics | grep db_connections
```

### High Memory Usage

```bash
# Check memory metrics
curl http://localhost:8080/metrics | grep memory

# Generate heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### Performance Issues

```bash
# Check request latency
curl http://localhost:8080/metrics | grep request_duration

# Generate CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

## Backup and Recovery

### Database Backup

```bash
# PostgreSQL
pg_dump -h db.example.com -U app_user production_db > backup.sql

# MySQL
mysqldump -h db.example.com -u app_user -p production_db > backup.sql
```

### Application State

```bash
# Backup configuration
tar -czf config-backup.tar.gz config/

# Backup logs
tar -czf logs-backup.tar.gz /var/log/app/
```

## Scaling

### Horizontal Scaling

Run multiple instances behind a load balancer:

```bash
# Instance 1
./app -addr :8080 -config config.yaml

# Instance 2
./app -addr :8081 -config config.yaml

# Instance 3
./app -addr :8082 -config config.yaml
```

### Vertical Scaling

Adjust resource limits:

```yaml
# config.yaml
database:
  max_open_conns: 50  # Increase for more throughput
  max_idle_conns: 10

cache:
  max_size: 1073741824  # 1GB
```

## Conclusion

This deployment guide covers the essential steps for deploying Rockstar Web Framework applications to production. Always test deployments in a staging environment first and monitor applications closely after deployment.

For additional help, consult the [documentation](.) or open an issue on GitHub.
