# K8s Scalable Web Application - Implementation Plan

## Project Name: `k8s-autoscale-webapp`

## GitHub Repository: `kubernetes-autoscale-webapp`

## Overview

This document provides a comprehensive implementation plan for deploying a 3-tier web application on Kubernetes with horizontal pod autoscaling. The application consists of a React frontend, Go API backend, PostgreSQL database, and Redis cache.

## Architecture Overview

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

## Directory Structure

```
kubernetes-autoscale-webapp/
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── public/
│   └── src/
│       ├── App.js
│       ├── components/
│       └── services/
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── handlers/
│   ├── models/
│   └── config/
├── k8s/
│   ├── namespace.yaml
│   ├── configmaps/
│   │   └── backend-config.yaml
│   ├── secrets/
│   │   └── db-credentials.yaml
│   ├── backend/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── hpa.yaml
│   ├── frontend/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── database/
│   │   ├── statefulset.yaml
│   │   ├── service.yaml
│   │   └── pvc.yaml
│   ├── redis/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── ingress/
│   │   └── ingress.yaml
│   └── monitoring/
│       └── metrics-server.yaml
├── scripts/
│   ├── deploy.sh
│   ├── load-test.sh
│   └── cleanup.sh
└── README.md
```

## Implementation Steps

### Step 1: Create the Backend Application (Go)

**File: backend/main.go**

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "github.com/go-redis/redis/v8"
    "context"
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB
var rdb *redis.Client
var ctx = context.Background()

func main() {
    // Initialize database connection
    initDB()
    defer db.Close()

    // Initialize Redis connection
    initRedis()
    defer rdb.Close()

    // Setup routes
    router := mux.NewRouter()

    // Health check endpoint
    router.HandleFunc("/health", healthHandler).Methods("GET")

    // API endpoints
    router.HandleFunc("/api/users", getUsers).Methods("GET")
    router.HandleFunc("/api/users", createUser).Methods("POST")
    router.HandleFunc("/api/users/{id}", getUser).Methods("GET")
    router.HandleFunc("/api/stress", stressTest).Methods("GET")

    // Enable CORS
    router.Use(corsMiddleware)

    log.Println("Server starting on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", router))
}

func initDB() {
    // Connection string from environment variables
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"))

    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    // Create users table if not exists
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100),
        email VARCHAR(100) UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

    _, err = db.Exec(createTableQuery)
    if err != nil {
        log.Fatal(err)
    }
}

func initRedis() {
    rdb = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_HOST") + ":6379",
        Password: "",
        DB:       0,
    })

    _, err := rdb.Ping(ctx).Result()
    if err != nil {
        log.Printf("Redis connection failed: %v", err)
    }
}

func stressTest(w http.ResponseWriter, r *http.Request) {
    // CPU intensive operation for testing HPA
    iterations := 100000000
    result := 0
    for i := 0; i < iterations; i++ {
        result += i
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Stress test completed",
        "result": result,
    })
}
```

**File: backend/Dockerfile**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Step 2: Create the Frontend Application (React)

**File: frontend/src/App.js**

```javascript
import React, { useState, useEffect } from 'react'
import axios from 'axios'
import './App.css'

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080'

function App() {
  const [users, setUsers] = useState([])
  const [newUser, setNewUser] = useState({ name: '', email: '' })
  const [loading, setLoading] = useState(false)
  const [stressTestResult, setStressTestResult] = useState('')

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    try {
      const response = await axios.get(`${API_URL}/api/users`)
      setUsers(response.data || [])
    } catch (error) {
      console.error('Error fetching users:', error)
    }
  }

  const createUser = async (e) => {
    e.preventDefault()
    try {
      await axios.post(`${API_URL}/api/users`, newUser)
      setNewUser({ name: '', email: '' })
      fetchUsers()
    } catch (error) {
      console.error('Error creating user:', error)
    }
  }

  const runStressTest = async () => {
    setLoading(true)
    setStressTestResult('Running stress test...')
    try {
      const response = await axios.get(`${API_URL}/api/stress`)
      setStressTestResult(
        'Stress test completed: ' + JSON.stringify(response.data)
      )
    } catch (error) {
      setStressTestResult('Stress test failed: ' + error.message)
    }
    setLoading(false)
  }

  return (
    <div className='App'>
      <h1>Kubernetes Auto-Scaling Demo</h1>

      <div className='section'>
        <h2>Create User</h2>
        <form onSubmit={createUser}>
          <input
            type='text'
            placeholder='Name'
            value={newUser.name}
            onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
            required
          />
          <input
            type='email'
            placeholder='Email'
            value={newUser.email}
            onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
            required
          />
          <button type='submit'>Create User</button>
        </form>
      </div>

      <div className='section'>
        <h2>Users</h2>
        <ul>
          {users.map((user) => (
            <li key={user.id}>
              {user.name} - {user.email}
            </li>
          ))}
        </ul>
      </div>

      <div className='section'>
        <h2>Load Testing</h2>
        <button onClick={runStressTest} disabled={loading}>
          {loading ? 'Running...' : 'Run Stress Test'}
        </button>
        {stressTestResult && <p>{stressTestResult}</p>}
      </div>
    </div>
  )
}

export default App
```

**File: frontend/Dockerfile**

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Step 3: Kubernetes Manifests

**File: k8s/namespace.yaml**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: webapp
```

**File: k8s/secrets/db-credentials.yaml**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
  namespace: webapp
type: Opaque
stringData:
  username: postgres
  password: postgres123
```

**File: k8s/configmaps/backend-config.yaml**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
  namespace: webapp
data:
  DB_HOST: 'postgres-service'
  DB_PORT: '5432'
  DB_NAME: 'webapp'
  REDIS_HOST: 'redis-service'
```

**File: k8s/database/statefulset.yaml**

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: webapp
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:15
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: username
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: password
            - name: POSTGRES_DB
              value: webapp
          volumeMounts:
            - name: postgres-storage
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              memory: '256Mi'
              cpu: '250m'
            limits:
              memory: '512Mi'
              cpu: '500m'
  volumeClaimTemplates:
    - metadata:
        name: postgres-storage
      spec:
        accessModes: ['ReadWriteOnce']
        resources:
          requests:
            storage: 5Gi
```

**File: k8s/database/service.yaml**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: webapp
spec:
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
  clusterIP: None
```

**File: k8s/redis/deployment.yaml**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: webapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
          resources:
            requests:
              memory: '128Mi'
              cpu: '100m'
            limits:
              memory: '256Mi'
              cpu: '200m'
```

**File: k8s/redis/service.yaml**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: webapp
spec:
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
```

**File: k8s/backend/deployment.yaml**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: webapp
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
        - name: backend
          image: backend:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
          env:
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: username
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: password
          envFrom:
            - configMapRef:
                name: backend-config
          resources:
            requests:
              memory: '128Mi'
              cpu: '100m'
            limits:
              memory: '256Mi'
              cpu: '200m'
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 10
```

**File: k8s/backend/service.yaml**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: webapp
spec:
  selector:
    app: backend
  ports:
    - port: 8080
      targetPort: 8080
```

**File: k8s/backend/hpa.yaml**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: webapp
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 50
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 70
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 30
      policies:
        - type: Percent
          value: 50
          periodSeconds: 30
        - type: Pods
          value: 2
          periodSeconds: 60
```

**File: k8s/frontend/deployment.yaml**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: webapp
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
        - name: frontend
          image: frontend:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
          env:
            - name: REACT_APP_API_URL
              value: 'http://backend-service:8080'
          resources:
            requests:
              memory: '64Mi'
              cpu: '50m'
            limits:
              memory: '128Mi'
              cpu: '100m'
```

**File: k8s/frontend/service.yaml**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: webapp
spec:
  selector:
    app: frontend
  ports:
    - port: 80
      targetPort: 80
  type: ClusterIP
```

**File: k8s/ingress/ingress.yaml**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: webapp-ingress
  namespace: webapp
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
    - host: webapp.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: frontend-service
                port:
                  number: 80
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: backend-service
                port:
                  number: 8080
```

### Step 4: Deployment Scripts

**File: scripts/deploy.sh**

```bash
#!/bin/bash
set -e

echo "Building Docker images..."
# Build backend
cd backend
docker build -t backend:latest .
cd ..

# Build frontend
cd frontend
docker build -t frontend:latest .
cd ..

echo "Deploying to Kubernetes..."
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Deploy metrics server for HPA
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Wait for metrics server
echo "Waiting for metrics-server to be ready..."
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=60s

# Deploy configurations
kubectl apply -f k8s/secrets/
kubectl apply -f k8s/configmaps/

# Deploy database
kubectl apply -f k8s/database/

# Wait for database to be ready
echo "Waiting for database to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n webapp --timeout=120s

# Deploy Redis
kubectl apply -f k8s/redis/

# Deploy backend
kubectl apply -f k8s/backend/

# Deploy frontend
kubectl apply -f k8s/frontend/

# Deploy ingress
kubectl apply -f k8s/ingress/

echo "Deployment completed!"
echo "Add '127.0.0.1 webapp.local' to your /etc/hosts file"
echo "Access the application at http://webapp.local"
```

**File: scripts/load-test.sh**

```bash
#!/bin/bash

echo "Starting load test..."
echo "This will generate CPU load to trigger HPA"

# Get the backend service endpoint
BACKEND_URL="http://webapp.local/api/stress"

# Run concurrent requests
echo "Sending 100 concurrent requests..."
for i in {1..100}; do
  curl -s $BACKEND_URL &
done

wait

echo "Load test completed!"
echo "Check HPA status with: kubectl get hpa -n webapp -w"
```

**File: scripts/cleanup.sh**

```bash
#!/bin/bash

echo "Cleaning up resources..."
kubectl delete namespace webapp
echo "Cleanup completed!"
```

### Step 5: Monitoring and Verification Commands

```bash
# Watch HPA status
kubectl get hpa -n webapp -w

# Check pod scaling
kubectl get pods -n webapp -w

# View resource usage
kubectl top pods -n webapp

# Check backend logs
kubectl logs -n webapp -l app=backend -f

# Describe HPA for detailed status
kubectl describe hpa backend-hpa -n webapp

# Port forward to access without ingress
kubectl port-forward -n webapp svc/frontend-service 3000:80
kubectl port-forward -n webapp svc/backend-service 8080:8080
```

## Testing Auto-Scaling

1. Deploy the application using `./scripts/deploy.sh`
2. Monitor HPA: `kubectl get hpa -n webapp -w`
3. Run load test: `./scripts/load-test.sh`
4. Watch pods scale up: `kubectl get pods -n webapp -w`
5. Stop load test and watch pods scale down after stabilization window

## Key Learning Points

1. **Resource Requests/Limits**: Essential for HPA to calculate utilization
2. **Metrics Server**: Required for resource-based HPA
3. **Stabilization Windows**: Prevent flapping during scaling
4. **Health Probes**: Ensure pods are ready before receiving traffic
5. **ConfigMaps/Secrets**: Proper configuration management
6. **StatefulSet vs Deployment**: Use StatefulSet for databases
7. **Service Discovery**: Internal DNS for service communication

## Prerequisites for Implementation

1. Kubernetes cluster (minikube, kind, or cloud provider)
2. kubectl configured
3. Docker installed
4. Ingress controller (nginx-ingress)
5. Metrics server for HPA

## Expected Outcomes

- Backend scales from 2 to 10 pods based on CPU usage
- Frontend remains at 2 replicas (no HPA)
- Database and Redis maintain single instances
- Zero-downtime during scaling events
- Proper resource utilization tracking

## Local Development

1. Use `kind` (https://kind.sigs.k8s.io/) tool to for local testing and deployment
2. Use `taskfile.yaml` extensively for all development commands (https://taskfile.dev/)

- Use Latest Go version, React and other tools (all latest versions)
