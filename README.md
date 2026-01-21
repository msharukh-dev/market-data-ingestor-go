# WS Ingestor

A high-performance Go application for ingesting real-time market data from WebSocket streams. It processes incoming data in batches, stores it in PostgreSQL for persistence, and caches it in Redis for fast access.

## Features

- **Real-time Data Ingestion**: Connects to WebSocket endpoints and receives live market data streams
- **High Performance**: Worker pool pattern with configurable concurrency for efficient batch processing
- **Persistent Storage**: PostgreSQL backend for reliable data persistence with dynamic table creation
- **Caching Layer**: Redis integration for fast in-memory access to recent market data
- **Metrics & Monitoring**: Prometheus metrics exposed on dedicated endpoint for observability
- **Graceful Shutdown**: Proper signal handling to ensure clean shutdown and data consistency
- **Health Checks**: Built-in health check endpoint for monitoring application status

## Architecture

```
WebSocket Client
      ↓
  Processor (Worker Pool)
      ↓
  ┌───┴───┐
  ↓       ↓
 Redis  PostgreSQL
```

### Components

- **WebSocket Client** (`internal/app/services/websocket/`): Connects to WebSocket endpoints and streams market data
- **Processor** (`cmd/processor/`): Worker pool implementation for concurrent data processing
- **Storage Layer** (`internal/app/services/storage/`):
  - PostgreSQL: Persistent storage with batch insert support
  - Redis: In-memory cache for quick retrieval
- **Configuration** (`internal/app/config/`): Environment-based configuration management
- **Logger** (`internal/app/common/logger/`): Structured logging with Logrus
- **Models** (`internal/app/models/`): Data structures for market data

## Prerequisites

- **Go**: Version 1.25.5 or higher
- **PostgreSQL**: 12 or later
- **Redis**: 6 or later
- **WebSocket Feed**: Access to a market data WebSocket endpoint

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/msharukh-dev/market-data-ingestor-go.git
cd market-data-ingestor-go
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Build the Application

```bash
go build -o ws_ingestor ./cmd/app
```

## Configuration

Create a `.env` file in the project root with the following variables:

### Required Variables

```env
WS_URL=ws://your-websocket-endpoint:port/ws
WS_API_KEY=your-api-key-here
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

### Optional Variables

```env
# Processing
WORKER_COUNT=10
BATCH_SIZE=100

# Database
MARKET_DATA_TABLE_NAME=market_data
API_KEYS_TABLE_NAME=api_keys

# Redis
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

# Server
WS_SERVER_ADDR=127.0.0.1:8080
APP_ENV=local
APP_NAME=market-data-ingestor
```

### Environment Variables Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `WS_URL` | WebSocket endpoint URL | Required |
| `WS_API_KEY` | API key for WebSocket authentication | Required |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `WORKER_COUNT` | Number of worker goroutines | 10 |
| `BATCH_SIZE` | Number of records per batch | 100 |
| `MARKET_DATA_TABLE_NAME` | PostgreSQL table for market data | market_data |
| `REDIS_ADDR` | Redis server address | 127.0.0.1:6379 |
| `REDIS_PASSWORD` | Redis password (if required) | Empty |
| `WS_SERVER_ADDR` | Internal WebSocket server address | 127.0.0.1:8080 |

## Usage

### Running the Application

```bash
# Run with environment variables from .env
./ws_ingestor

# Or specify environment variables directly
WS_URL=ws://... WS_API_KEY=... DATABASE_URL=... ./ws_ingestor
```

### Health Check

The application exposes a health check endpoint:

```bash
curl http://localhost:8080/health
```

### Metrics

Prometheus metrics are available at:

```bash
curl http://localhost:8080/metrics
```

## Project Structure

```
ws_ingestor/
├── cmd/
│   ├── app/
│   │   ├── main.go           # Application entry point
│   │   └── bootstrap/        # Initialization logic
│   └── processor/            # Batch processing logic
├── internal/
│   ├── app/
│   │   ├── common/           # Shared utilities
│   │   ├── config/           # Configuration management
│   │   ├── constants/        # Application constants
│   │   ├── dto/              # Data transfer objects
│   │   ├── models/           # Data structures
│   │   ├── metrics/          # Prometheus metrics
│   │   └── services/         # Business logic
│   └── config/               # Alternative config location
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── .env                      # Environment configuration
└── README.md                 # This file
```

## Development

### Building from Source

```bash
# Build for current OS
go build -o ws_ingestor ./cmd/app

# Build for Linux (from Windows)
GOOS=linux GOARCH=amd64 go build -o ws_ingestor ./cmd/app

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ws_ingestor ./cmd/app
```

### Running Tests

```bash
go test ./...
```

### Code Organization

- **cmd/**: Executable applications
- **internal/**: Private application code
- **internal/app/**: Application-specific logic
- **internal/utils/**: Utility functions

## How It Works

1. **Data Ingestion**: WebSocket client connects to the market data feed and receives messages
2. **Buffering**: Incoming data is buffered in a channel for concurrent processing
3. **Batch Processing**: Worker pool processes data in configurable batch sizes
4. **Storage**: Processed data is simultaneously stored in PostgreSQL (persistent) and Redis (cache)
5. **Metrics**: Application metrics are collected and exposed for monitoring
6. **Graceful Shutdown**: On termination signal, the application waits for in-flight operations to complete

## Performance Characteristics

- **Throughput**: Configurable worker count and batch size for optimized processing
- **Latency**: Redis caching reduces read latency for recent data
- **Storage**: PostgreSQL batch inserts minimize database load
- **Concurrency**: Worker pool pattern enables horizontal scaling through configuration

## Monitoring

The application exposes Prometheus metrics on the health/metrics endpoint:

```bash
curl http://127.0.0.1:8080/metrics
```

Metrics include:
- Data ingestion counts
- Processing latency
- Storage operation metrics
- WebSocket connection status

## Error Handling

The application includes comprehensive error handling:
- Graceful recovery from WebSocket disconnections
- Connection pooling for database resilience
- Panic recovery with detailed logging
- Error propagation for operational visibility

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository: https://github.com/msharukh-dev/market-data-ingestor-go
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or suggestions:

1. Check existing GitHub issues
2. Create a new issue with detailed information
3. Include relevant logs and configuration (without sensitive data)

## Roadmap

- [ ] Add comprehensive test coverage
- [ ] Support multiple WebSocket feeds simultaneously
- [ ] Database migration tooling
- [ ] Docker containerization
- [ ] Kubernetes deployment manifests
- [ ] Advanced filtering and aggregation options

## Changelog

### Version 1.0.0
- Initial release
- WebSocket data ingestion
- PostgreSQL persistence
- Redis caching
- Worker pool processing
- Prometheus metrics
- Health check endpoints
