# Saga Account
Account service of the [saga pattern implementation](https://github.com/minghsu0107/saga-example).

Features:
- High performance gRPC authentication
- JWT token management
- Caching middleware proxy compatible with repository interface
- Local + Redis cache
- Request coalescing to prevent cache avalanche
- Prometheus metrics
- Distributed tracing exporter
  - HTTP server 
  - gPRC server
- Comprehensive application struture with domain-driven design (DDD), decoupling service implementations from configurations and transports
- Compile-time dependecy injection using [wire](https://github.com/google/wire)
- Graceful shutdown
- Unit testing and continuous integration using [Drone CI](https://www.drone.io)

## Usage
Setup githooks:
```bash=
git config core.hooksPath githooks
```
Build from source:
```bash
make build
```
Start the service:
```bash
DB_DSN="ming:password@tcp(accountdb:3306)/account?charset=utf8mb4&parseTime=True&loc=Local" \
REDIS_ADDRS=redis-node1:7000,redis-node2:7001,redis-node3:7002,redis-node4:7003,redis-node5:7004,redis-node6:7005 \
REDIS_PASSWORD=myredispassword \
JWT_ACCESS_TOKEN_EXPIRE_SECOND=10800 \
JWT_REFRESH_TOKEN_EXPIRE_SECOND=86400 \
JAEGER_URL=http://jaeger:14268/api/traces \
./server
```
Test locally:
```bash
DB_DSN="ming:password@tcp(<mysql-host>:3306)/account?charset=utf8mb4&parseTime=True&loc=Local" \
make test
```
- `DB_DSN`: MySQL connection DSN.
- `REDIS_ADDRS`: Redis seed server addresses
- `JWT_ACCESS_TOKEN_EXPIRE_SECOND`: access token expiration duration (second)
- `JWT_REFRESH_TOKEN_EXPIRE_SECOND`: refresh token expiration duration (second)
## Running in Docker
See [docker-compose example](https://github.com/minghsu0107/saga-example/blob/main/docker-compose.yaml) for details.
## Exported Metrics
| Metric                                                                                                                               | Description                                                                                 | Labels                      |
| ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------- | --------------------------- |
| account_http_request_duration_seconds (account_http_request_duration_seconds_count, account_http_request_duration_seconds_bucket, account_http_request_duration_sum) | A Prometheus histogram. Records the latency of the HTTP requests.                           | `code`, `handler`, `method` |
| account_http_requests_inflight                                                                                                               | A Prometheus gauge. Records the number of inflight requests being handled at the same time. | `code`, `handler`, `method` |
| account_http_response_size_bytes (account_http_response_size_bytes_count, account_http_response_size_bytes_bucket, account_http_response_size_bytes_sum)             | A Prometheus histogram. Records the size of the HTTP responses.                             | `handler`                   |
