# Backend API Examples and Expected Responses

This document provides example responses for the REST API endpoints defined in `backend-api.rest`.

## Prerequisites

1. Install [REST Client Extension](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) in VS Code
2. Ensure backend is running with port forwarding:
   ```bash
   task port:backend
   # or
   kubectl port-forward -n webapp svc/backend-service 8080:8080
   ```

## API Endpoints

### 1. Health Check
**Request:** `GET /health`

**Expected Response:**
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "timestamp": "2025-07-19T16:00:10.728058189Z"
}
```

### 2. Get All Users
**Request:** `GET /api/users`

**Expected Response (Empty):**
```json
null
```

**Expected Response (With Users):**
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "created_at": "2025-07-19T16:00:18.373599Z"
  },
  {
    "id": 2,
    "name": "Jane Smith",
    "email": "jane.smith@example.com",
    "created_at": "2025-07-19T16:01:25.123456Z"
  }
]
```

### 3. Create User
**Request:** `POST /api/users`
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com"
}
```

**Expected Response:**
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john.doe@example.com",
  "created_at": "2025-07-19T16:00:18.373599Z"
}
```

**Error Response (Duplicate Email):**
```
pq: duplicate key value violates unique constraint "users_email_key"
```

### 4. Get User by ID
**Request:** `GET /api/users/{id}`

**Expected Response (Found):**
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john.doe@example.com",
  "created_at": "2025-07-19T16:00:18.373599Z"
}
```

**Error Response (Not Found):**
```
User not found
```

### 5. Stress Test
**Request:** `GET /api/stress`

**Expected Response:**
```json
{
  "message": "Stress test completed",
  "result": 4999999950000000,
  "iterations": 100000000
}
```

## Using REST Client Features

### 1. Send Request
- Click on "Send Request" text above any request
- Or use `Cmd/Ctrl + Alt + R` when cursor is on a request

### 2. View Response
- Response appears in a split panel
- Shows status code, headers, and body
- Response time is displayed

### 3. Request History
- Use `Cmd/Ctrl + Alt + H` to view history
- Rerun previous requests easily

### 4. Environment Variables
- Change `@baseUrl` at the top of the file
- Supports different environments (localhost, kubernetes service, etc.)

### 5. Dynamic Values
- Use `{{requestName.response.body.field}}` to reference previous responses
- Example: `{{createUser.response.body.id}}` gets the ID from a previous create

### 6. Save Responses
- Right-click on response panel
- Select "Save Response Body"

## Testing Scenarios

### 1. Basic CRUD Flow
```
1. Send: Get All Users (should be empty/null)
2. Send: Create New User
3. Send: Get All Users (should show the created user)
4. Send: Get User by ID (use the ID from step 2)
```

### 2. Cache Testing
```
1. Send: Get All Users (hits database)
2. Immediately Send: Get All Users again (should be cached - faster)
3. Send: Create New User
4. Send: Get All Users (cache cleared, hits database)
```

### 3. Load Testing for HPA
```
1. Open terminal: watch 'kubectl get hpa -n webapp'
2. Send: Multiple Stress Test requests rapidly
3. Watch HPA scale up pods based on CPU usage
4. Stop sending requests
5. Watch HPA scale down after cooldown period
```

### 4. Error Handling
```
1. Send: Create User with duplicate email
2. Send: Get non-existent user (ID: 99999)
3. Send: Malformed JSON request
4. Send: Empty body request
```

## Monitoring During Tests

### Watch HPA Status
```bash
kubectl get hpa -n webapp -w
```

### Watch Pod Scaling
```bash
kubectl get pods -n webapp -w
```

### Monitor Resource Usage
```bash
task top:pods
# or
kubectl top pods -n webapp
```

### View Logs
```bash
task logs:backend
# or
kubectl logs -n webapp -l app=backend -f
```

## Troubleshooting

### Connection Refused
- Ensure port forwarding is active: `task port:backend`
- Check if backend pods are running: `kubectl get pods -n webapp`

### 500 Internal Server Error
- Check backend logs: `task logs:backend`
- Verify database connection: `task logs:postgres`

### Slow Responses
- Check if Redis is working: `task logs:redis`
- Monitor CPU/memory usage: `task top:pods`

## Performance Notes

- First request after startup may be slower (cold start)
- Cached requests (GET /api/users) return much faster
- Stress endpoint intentionally uses CPU to trigger scaling
- Database queries are not optimized for large datasets