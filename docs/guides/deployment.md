# Deployment Guide

## Overview

This guide covers deploying Rockstar Web Framework applications to production environments. Learn about containerization, configuration management, monitoring setup, scaling strategies, and troubleshooting production issues.

**Deployment topics covered:**
- **Containerization**: Docker and Kubernetes deployment
- **Configuration**: Environment-specific settings
- **Monitoring**: Production observability
- **Scaling**: Horizontal and vertical scaling
- **Security**: Production security hardening
- **CI/CD**: Automated deployment pipelines
- **Troubleshooting**: Common production issues

## Quick Start

Basic production deployment:

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o rockstar-app cmd/rockstar/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rockstar-app .
COPY --from=builder /app/config.production.yaml ./config.yaml
EXPOSE 8080
CMD ["./rockstar-app"]
```

```bash
# Build and run
docker build -t rockstar-app .
docker run -p 8080:8080 rockstar-app
```

## Containerization with Docker

### Production Dockerfile

Create an optimized production Dockerfile:

```dockerfile
# Multi-stage build for smaller image
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o rockstar-app \
    cmd/rockstar/main.go

# Production image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/rockstar-app .

# Copy configuration and assets
COPY --from=builder /app/config.production.yaml ./config.yaml
COPY --from=builder /app/sql ./sql
COPY --from=builder /app/locales ./locales

# Set ownership
RUN chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090 6060

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
CMD ["./rockstar-app"]
```

### Docker Compose

Create a docker-compose.yml for local development and testing:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"  # Metrics
      - "6060:6060"  # Pprof
    environment:
      - ROCKSTAR_ENV=production
      - ROCKSTAR_DATABASE_HOST=postgres
      - ROCKSTAR_DATABASE_PASSWORD=${DB_PASSWORD}
    depends_on:
      - postgres
      - redis
    volumes:
      - ./logs:/home/appuser/logs
    restart: unless-stopped
    networks:
      - app-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=rockstar
      - POSTGRES_USER=rockstar
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped
    networks:
      - app-network

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    restart: unless-stopped
    networks:
      - app-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana-data:/var/lib/grafana
    restart: unless-stopped
    networks:
      - app-network

volumes:
  postgres-data:
  redis-data:
  prometheus-data:
  grafana-data:

networks:
  app-network:
    driver: bridge
```

### Building and Running

```bash
# Build image
docker build -t rockstar-app:latest .

# Run container
docker run -d \
  --name rockstar-app \
  -p 8080:8080 \
  -e ROCKSTAR_ENV=production \
  -e ROCKSTAR_DATABASE_HOST=postgres \
  -e ROCKSTAR_DATABASE_PASSWORD=secret \
  rockstar-app:latest

# View logs
docker logs -f rockstar-app

# Stop container
docker stop rockstar-app

# Remove container
docker rm rockstar-app
```

## Kubernetes Deployment

### Deployment Manifest

Create a Kubernetes deployment:

```yaml
# deployment.yaml
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
      - name: rockstar-app
        image: rockstar-app:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: ROCKSTAR_ENV
          value: "production"
        - name: ROCKSTAR_DATABASE_HOST
          valueFrom:
            configMapKeyRef:
              name: rockstar-config
              key: database.host
        - name: ROCKSTAR_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rockstar-secrets
              key: database.password
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /home/appuser/config.yaml
          subPath: config.yaml
      volumes:
      - name: config
        configMap:
          name: rockstar-config
---
apiVersion: v1
kind: Service
metadata:
  name: rockstar-app
spec:
  selector:
    app: rockstar-app
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer
```

### ConfigMap

Store configuration in ConfigMap:

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rockstar-config
data:
  database.host: "postgres-service"
  database.port: "5432"
  database.name: "rockstar"
  config.yaml: |
    server:
      http1_enabled: true
      http2_enabled: true
      max_connections: 10000
    monitoring:
      enable_metrics: true
      metrics_port: 9090
    cache:
      type: memory
      max_size: 104857600
```

### Secrets

Store sensitive data in Secrets:

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: rockstar-secrets
type: Opaque
stringData:
  database.password: "your-secure-password"
  jwt.secret: "your-jwt-secret"
  session.key: "your-32-byte-session-key"
```

### Horizontal Pod Autoscaler

Auto-scale based on metrics:

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rockstar-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rockstar-app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Ingress

Configure ingress for external access:

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rockstar-app-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.example.com
    secretName: rockstar-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: rockstar-app
            port:
              number: 80
```

### Deploying to Kubernetes

```bash
# Create namespace
kubectl create namespace rockstar

# Apply configurations
kubectl apply -f configmap.yaml -n rockstar
kubectl apply -f secrets.yaml -n rockstar
kubectl apply -f deployment.yaml -n rockstar
kubectl apply -f hpa.yaml -n rockstar
kubectl apply -f ingress.yaml -n rockstar

# Check deployment status
kubectl get deployments -n rockstar
kubectl get pods -n rockstar
kubectl get services -n rockstar

# View logs
kubectl logs -f deployment/rockstar-app -n rockstar

# Scale manually
kubectl scale deployment rockstar-app --replicas=5 -n rockstar

# Update deployment
kubectl set image deployment/rockstar-app \
  rockstar-app=rockstar-app:v2 -n rockstar

# Rollback deployment
kubectl rollout undo deployment/rockstar-app -n rockstar
```

## Production Configuration

### Environment-Specific Configuration

Create environment-specific config files:

```yaml
# config.production.yaml
server:
  http1_enabled: true
  http2_enabled: true
  read_timeout: 30s
  write_timeout: 30s
  max_connections: 10000

database:
  driver: postgres
  host: ${ROCKSTAR_DATABASE_HOST}
  port: 5432
  database: rockstar_prod
  username: rockstar
  password: ${ROCKSTAR_DATABASE_PASSWORD}
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 10m

cache:
  type: memory
  max_size: 104857600  # 100MB
  default_ttl: 5m

session:
  encryption_key: ${ROCKSTAR_SESSION_KEY}
  cookie_secure: true
  cookie_http_only: true
  cookie_same_site: Strict
  session_lifetime: 24h

monitoring:
  enable_metrics: true
  enable_pprof: false  # Disable in production
  metrics_port: 9090
  enable_optimization: true
  optimization_interval: 5m

security:
  enable_csrf: true
  enable_xss_protect: true
  enable_hsts: true
  hsts_max_age: 31536000
```

### Environment Variables

Use environment variables for sensitive data:

```bash
# .env.production
ROCKSTAR_ENV=production
ROCKSTAR_DATABASE_HOST=postgres.example.com
ROCKSTAR_DATABASE_PASSWORD=secure-password
ROCKSTAR_SESSION_KEY=your-32-byte-encryption-key
ROCKSTAR_JWT_SECRET=your-jwt-secret
```

### Configuration Loading

Load configuration in your application:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
    "os"
)

func main() {
    // Load environment-specific config
    env := os.Getenv("ROCKSTAR_ENV")
    if env == "" {
        env = "development"
    }
    
    configFile := fmt.Sprintf("config.%s.yaml", env)
    
    config := pkg.NewConfigManager()
    if err := config.Load(configFile); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Override with environment variables
    if err := config.LoadFromEnv(); err != nil {
        log.Fatalf("Failed to load env vars: %v", err)
    }
    
    // Create framework config
    frameworkConfig := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP1:    config.GetBool("server.http1_enabled"),
            EnableHTTP2:    config.GetBool("server.http2_enabled"),
            MaxConnections: config.GetInt("server.max_connections"),
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   config.GetString("database.driver"),
            Host:     config.GetString("database.host"),
            Port:     config.GetInt("database.port"),
            Database: config.GetString("database.database"),
            Username: config.GetString("database.username"),
            Password: config.GetString("database.password"),
        },
        // ... other config
    }
    
    app, err := pkg.New(frameworkConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Fatal(app.Listen(":8080"))
}
```

## Monitoring and Logging

### Prometheus Setup

Configure Prometheus to scrape metrics:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'rockstar-app'
    static_configs:
      - targets: ['rockstar-app:9090']
        labels:
          environment: 'production'
          app: 'rockstar-app'
    
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

### Grafana Dashboards

Create Grafana dashboards for visualization:

```json
{
  "dashboard": {
    "title": "Rockstar App Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests[5m])"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_errors[5m])"
          }
        ]
      },
      {
        "title": "Response Time (P95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, http_request_duration)"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "targets": [
          {
            "expr": "system_memory_usage"
          }
        ]
      }
    ]
  }
}
```

### Structured Logging

Implement structured logging:

```go
import (
    "log/slog"
    "os"
)

func main() {
    // Create JSON logger for production
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    
    slog.SetDefault(logger)
    
    // Log with structured fields
    logger.Info("Application starting",
        "version", "1.0.0",
        "environment", "production",
        "port", 8080,
    )
    
    // ... application code
}

// In handlers
router.GET("/api/users", func(ctx pkg.Context) error {
    logger := ctx.Logger()
    
    logger.Info("Fetching users",
        "tenant_id", ctx.Tenant().ID,
        "user_id", ctx.User().ID,
    )
    
    // ... handler logic
    
    return ctx.JSON(200, users)
})
```

### Log Aggregation

Send logs to centralized logging:

```yaml
# fluentd-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      format json
    </source>
    
    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      logstash_format true
      logstash_prefix rockstar-app
    </match>
```

## Scaling Strategies

### Horizontal Scaling

Scale by adding more instances:

```bash
# Kubernetes
kubectl scale deployment rockstar-app --replicas=10

# Docker Swarm
docker service scale rockstar-app=10

# Manual with load balancer
# Deploy multiple instances behind load balancer
```

**When to scale horizontally:**
- High request volume
- CPU-bound workloads
- Need for high availability
- Geographic distribution

### Vertical Scaling

Scale by increasing resources:

```yaml
# Increase resource limits
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

**When to scale vertically:**
- Memory-intensive operations
- Single-threaded bottlenecks
- Database connection limits
- Before horizontal scaling

### Auto-Scaling

Configure automatic scaling:

```yaml
# Kubernetes HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rockstar-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rockstar-app
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1000"
```

### Load Balancing

Configure load balancer:

```nginx
# nginx.conf
upstream rockstar_app {
    least_conn;  # Load balancing method
    
    server app1:8080 max_fails=3 fail_timeout=30s;
    server app2:8080 max_fails=3 fail_timeout=30s;
    server app3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.example.com;
    
    location / {
        proxy_pass http://rockstar_app;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Health checks
        proxy_next_upstream error timeout http_502 http_503 http_504;
    }
    
    location /health {
        access_log off;
        proxy_pass http://rockstar_app/health;
    }
}
```

## Security Hardening

### TLS/SSL Configuration

Enable HTTPS in production:

```go
config := pkg.ServerConfig{
    TLSEnabled:  true,
    TLSCertFile: "/etc/ssl/certs/server.crt",
    TLSKeyFile:  "/etc/ssl/private/server.key",
    EnableHTTP2: true,
}
```

### Security Headers

Configure security headers:

```go
router.Use(func(ctx pkg.Context) error {
    // Security headers
    ctx.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    ctx.SetHeader("X-Frame-Options", "DENY")
    ctx.SetHeader("X-Content-Type-Options", "nosniff")
    ctx.SetHeader("X-XSS-Protection", "1; mode=block")
    ctx.SetHeader("Content-Security-Policy", "default-src 'self'")
    ctx.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
    
    return ctx.Next()
})
```

### Secrets Management

Use secrets management tools:

```bash
# Kubernetes Secrets
kubectl create secret generic rockstar-secrets \
  --from-literal=database-password=secret \
  --from-literal=jwt-secret=secret

# HashiCorp Vault
vault kv put secret/rockstar \
  database-password=secret \
  jwt-secret=secret

# AWS Secrets Manager
aws secretsmanager create-secret \
  --name rockstar/database-password \
  --secret-string secret
```

### Network Policies

Restrict network access:

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: rockstar-app-policy
spec:
  podSelector:
    matchLabels:
      app: rockstar-app
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: nginx-ingress
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

## CI/CD Pipeline

### GitHub Actions

Create a CI/CD pipeline:

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -v ./...
      
      - name: Run benchmarks
        run: go test -bench=. ./tests/

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker image
        run: docker build -t rockstar-app:${{ github.sha }} .
      
      - name: Push to registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push rockstar-app:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/rockstar-app \
            rockstar-app=rockstar-app:${{ github.sha }} \
            -n production
          
          kubectl rollout status deployment/rockstar-app -n production
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.21
  script:
    - go test -v ./...
    - go test -bench=. ./tests/

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/rockstar-app 
        rockstar-app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA 
        -n production
    - kubectl rollout status deployment/rockstar-app -n production
  only:
    - main
```

## Troubleshooting

### Application Won't Start

**Symptoms**: Container exits immediately

**Solutions**:
```bash
# Check logs
docker logs rockstar-app
kubectl logs deployment/rockstar-app

# Common issues:
# - Missing configuration file
# - Invalid database credentials
# - Port already in use
# - Missing environment variables
```

### High Memory Usage

**Symptoms**: OOMKilled errors

**Solutions**:
```bash
# Increase memory limits
kubectl set resources deployment rockstar-app \
  --limits=memory=1Gi

# Enable memory optimization
# In config: enable_optimization: true

# Check for memory leaks
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### Database Connection Issues

**Symptoms**: Connection refused errors

**Solutions**:
```bash
# Check database connectivity
kubectl exec -it deployment/rockstar-app -- \
  nc -zv postgres-service 5432

# Verify credentials
kubectl get secret rockstar-secrets -o yaml

# Check connection pool settings
# Increase max_open_conns if needed
```

### Slow Response Times

**Symptoms**: High latency

**Solutions**:
```bash
# Profile application
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Check metrics
curl http://localhost:9090/metrics

# Scale horizontally
kubectl scale deployment rockstar-app --replicas=10
```

### Pod Crashes

**Symptoms**: CrashLoopBackOff

**Solutions**:
```bash
# Check pod events
kubectl describe pod <pod-name>

# View logs
kubectl logs <pod-name> --previous

# Common causes:
# - Failed health checks
# - Panic in application code
# - Resource limits too low
# - Missing dependencies
```

## Best Practices

### 1. Use Health Checks

Implement proper health checks:

```go
router.GET("/healthz", func(ctx pkg.Context) error {
    return ctx.String(200, "OK")
})

router.GET("/readyz", func(ctx pkg.Context) error {
    // Check dependencies
    if err := ctx.Database().Ping(); err != nil {
        return ctx.String(503, "Database not ready")
    }
    return ctx.String(200, "Ready")
})
```

### 2. Graceful Shutdown

Handle shutdown signals:

```go
func main() {
    app, _ := pkg.New(config)
    
    // Start server in goroutine
    go func() {
        if err := app.Listen(":8080"); err != nil {
            log.Fatal(err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := app.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

### 3. Monitor Everything

Set up comprehensive monitoring:
- Application metrics
- System metrics
- Database metrics
- Error rates
- Response times

### 4. Use Secrets Management

Never hardcode secrets:
- Use environment variables
- Use Kubernetes Secrets
- Use external secrets managers
- Rotate secrets regularly

### 5. Implement Rate Limiting

Protect against abuse:

```go
router.Use(rateLimitMiddleware)
```

### 6. Enable CORS Properly

Configure CORS for production:

```go
corsConfig := pkg.CORSConfig{
    AllowOrigins:     []string{"https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           3600,
}
```

### 7. Use CDN for Static Assets

Serve static assets through CDN:
- Reduced latency
- Lower bandwidth costs
- Better scalability
- DDoS protection

### 8. Regular Backups

Backup critical data:
- Database backups
- Configuration backups
- Secrets backups
- Test restore procedures

## See Also

- [Configuration Guide](configuration.md) - Configuration management
- [Monitoring Guide](monitoring.md) - Production monitoring
- [Performance Guide](performance.md) - Performance optimization
- [Security Guide](security.md) - Security best practices

