# Kubernetes Auto-Scaling Web Application

A comprehensive 3-tier web application demonstrating Kubernetes horizontal pod autoscaling (HPA) with React frontend, Go backend, PostgreSQL database, and Redis cache.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│  React Frontend │────▶│   Go Backend    │────▶│   PostgreSQL    │
│   (Port 3000)   │     │   (Port 8080)   │     │   (Port 5432)   │
│                 │     │                 │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │                 │
                        │   Redis Cache   │
                        │   (Port 6379)   │
                        │                 │
                        └─────────────────┘
```

## Features

- **Auto-scaling**: Backend pods scale from 2 to 10 based on CPU/memory usage
- **Health Monitoring**: Comprehensive health checks and monitoring
- **Caching**: Redis integration for improved performance
- **Load Testing**: Built-in stress testing endpoints to trigger HPA
- **Development Tools**: Complete Taskfile.yaml for all operations

## Quick Start

### Prerequisites

- Docker Desktop
- kubectl
- kind (Kubernetes in Docker)
- task (optional, for using Taskfile commands)

### 1. Setup Cluster and Deploy

```bash
# Using Task (recommended)
task setup

# Or using scripts directly
./scripts/deploy.sh
```

### 2. Access the Application

Add to `/etc/hosts`:
```
127.0.0.1 webapp.local
```

Then access: http://webapp.local

### 3. Test Auto-scaling

```bash
# Monitor HPA
kubectl get hpa -n webapp -w

# Run load test
task test:load-intensive

# Watch pods scale up
kubectl get pods -n webapp -w
```

## Development Commands

This project uses [Task](https://taskfile.dev/) for development workflows:

```bash
# Show all available commands
task

# Build and deploy everything
task setup

# Monitor resources
task status:all
task top:pods

# Port forward services
task port:backend
task port:frontend

# View logs
task logs:backend
task logs:frontend

# Load testing
task test:stress
task test:load

# Restart deployments
task restart:backend
task restart:frontend

# Cleanup
task clean:all
task clean:cluster
```

## Directory Structure

```
kubernetes-autoscale-webapp/
├── frontend/                 # React application
│   ├── Dockerfile
│   ├── package.json
│   ├── nginx.conf
│   └── src/
├── backend/                  # Go API server
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── k8s/                     # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmaps/
│   ├── secrets/
│   ├── backend/
│   ├── frontend/
│   ├── database/
│   ├── redis/
│   └── ingress/
├── scripts/                 # Deployment scripts
│   ├── deploy.sh
│   ├── load-test.sh
│   ├── cleanup.sh
│   ├── sustained-load.sh
│   └── intensive-load.sh
├── Taskfile.yaml           # Development commands
└── README.md
```

## API Endpoints

### Backend API

- `GET /health` - Health check with database/Redis status
- `GET /api/users` - List all users (cached)
- `POST /api/users` - Create new user
- `GET /api/users/{id}` - Get user by ID (cached)
- `GET /api/stress` - CPU-intensive endpoint for load testing

### Frontend Features

- User creation form
- User list display
- Load testing button
- Real-time status updates

## Monitoring and Scaling

### HPA Configuration

- **Min Replicas**: 2
- **Max Replicas**: 10
- **CPU Target**: 50% of requests
- **Memory Target**: 70% of requests
- **Scale Up**: 50% increase every 30s (max 2 pods/60s)
- **Scale Down**: 10% decrease every 60s

### Monitoring Commands

```bash
# Watch HPA status
kubectl get hpa -n webapp -w

# Check resource usage
kubectl top pods -n webapp

# View scaling events
kubectl describe hpa backend-hpa -n webapp

# Monitor pod count
kubectl get pods -n webapp -w
```

## Load Testing

### Built-in Tests

```bash
# Single stress test
task test:stress

# Sustained load (20 concurrent)
task test:load

# Intensive load (50 concurrent)
task test:load-intensive
```

### Manual Testing

```bash
# Test user creation
curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}' \
  http://localhost:8080/api/users

# Trigger CPU load
curl http://localhost:8080/api/stress
```

## Configuration

### Environment Variables

Backend configuration via ConfigMap and Secrets:

- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port
- `DB_NAME`: Database name
- `DB_USER`: Database username (from secret)
- `DB_PASSWORD`: Database password (from secret)
- `REDIS_HOST`: Redis host

### Resource Limits

```yaml
Backend:
  requests: { memory: 128Mi, cpu: 100m }
  limits: { memory: 256Mi, cpu: 200m }

Frontend:
  requests: { memory: 64Mi, cpu: 50m }
  limits: { memory: 128Mi, cpu: 100m }
```

## Troubleshooting

### Common Issues

1. **Metrics Server Not Working**
   ```bash
   kubectl patch deployment metrics-server -n kube-system --type='json' \
     -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'
   ```

2. **Pods Not Scaling**
   - Check metrics server: `kubectl top pods -n webapp`
   - Verify HPA: `kubectl describe hpa backend-hpa -n webapp`
   - Ensure resource requests are set

3. **Database Connection Issues**
   - Restart backend: `task restart:backend`
   - Check postgres logs: `task logs:postgres`

4. **Port Forwarding Issues**
   ```bash
   pkill -f "kubectl port-forward"
   task port:backend
   ```

### Useful Debugging Commands

```bash
# Check all resources
task status:all

# View detailed pod info
kubectl describe pods -n webapp

# Check events
kubectl get events -n webapp --sort-by='.lastTimestamp'

# Test connectivity
kubectl exec -it -n webapp deployment/backend -- wget -qO- http://postgres-service:5432
```

## Advanced Features

### Custom Metrics

The HPA can be extended with custom metrics:

```yaml
metrics:
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1k"
```

### Multi-cluster Deployment

For production, consider:

- External database (RDS, Cloud SQL)
- Redis Cluster mode
- External load balancer
- SSL/TLS termination
- Monitoring with Prometheus/Grafana

## Performance Testing Results

Typical scaling behavior:

- **Scale Up**: ~30-60 seconds after load increase
- **Scale Down**: ~60-120 seconds after load decrease
- **Maximum Throughput**: ~1000 requests/second per pod
- **Database Connections**: Pooled, max 100 per pod

## Contributing

1. Fork the repository
2. Create feature branch
3. Test locally with `task setup`
4. Submit pull request

## License

MIT License - see LICENSE file for details.

---

**Built with ❤️ for learning Kubernetes auto-scaling**