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
Build from source:
```bash
go mod tidy
make build
```
Start the service:
```bash
DB_DSN="ming:password@tcp(accountdb:3306)/account?charset=utf8mb4&parseTime=True&loc=Local" \
REDIS_ADDRS=redis-node1:7000,redis-node2:7001,redis-node3:7002,redis-node4:7003,redis-node5:7004,redis-node6:7005 \
REDIS_PASSWORD=myredispassword \
JWT_ACCESS_TOKEN_EXPIRE_SECOND=10800 \
JWT_REFRESH_TOKEN_EXPIRE_SECOND=86400 \
OC_AGENT_HOST=oc-collector:55678 \
./server
```
- `DB_DSN`: MySQL connection DSN.
- `REDIS_ADDRS`: Redis seed server addresses
- `JWT_ACCESS_TOKEN_EXPIRE_SECOND`: access token expiration duration (second)
- `JWT_REFRESH_TOKEN_EXPIRE_SECOND`: refresh token expiration duration (second)
## Running in Docker
See [docker-compose example](https://github.com/minghsu0107/saga-example/blob/main/docker-compose.yaml) for details.
## Exported Metrics
| Metric                                                                                                                                   | Description                                                                                                 | Labels                                                           |
| ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| subscriber_messages_received_total                                                                                                       | A Prometheus Counter. Counts the number of messages obtained by the subscriber.                             | `acked` ("acked" or "nacked"), `handler_name`, `subscriber_name` |
| handler_execution_time_seconds (handler_execution_time_seconds_count, handler_execution_time_bucket, handler_execution_time_seconds_sum) | A Prometheus Histogram. Records the execution time of the handler function wrapped by the middleware.       | `handler_name`, `success` ("true" or "false")                    |
| publish_time_seconds (publish_time_seconds_count, publish_time_seconds_bucket, publish_time_seconds_sum)                                 | A Prometheus Histogram. Registers the time of execution of the Publish function of the decorated publisher. | `handler_name`, `success` ("true" or "false"), `publisher_name`  |
| http_request_duration_seconds (http_request_duration_seconds_count, http_request_duration_seconds_bucket, http_request_duration_sum)     | A Prometheus histogram. Records the latency of the HTTP requests.                                           | `code`, `handler`, `method`                                      |
| http_requests_inflight                                                                                                                   | A Prometheus gauge. Records the number of inflight requests being handled at the same time.                 | `code`, `handler`, `method`                                      |
| http_response_size_bytes (http_response_size_bytes_count, http_response_size_bytes_bucket, http_response_size_bytes_sum)                 | A Prometheus histogram. Records the size of the HTTP responses.                                             | `handler`                                                        |