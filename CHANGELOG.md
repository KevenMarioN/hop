# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Prometheus Metrics Integration**: Added optional metrics collection with `WithMetrics()` option
  - Metrics: `hop_messages_consumed_total`, `hop_consumption_errors_total`, `hop_reconnects_total`, `hop_connection_duration_seconds`, `hop_active_consumers`
- **ConsumerBuilder**: New fluent API for building immutable Consumer configurations via `protocol.NewConsumerBuilder()`
- **Improved Documentation**: Added comprehensive go doc comments to all public types and functions
- **HTTP Metrics Endpoint**: New `metrics.Handler()` for exposing Prometheus metrics

### Changed
- **Race Condition Fix**: Fixed potential race condition in `Manager.Start()` by copying consumer map before iteration
- **Performance Optimization**: Removed per-message debug logging from `Consumer.Execute()` (metrics provide observability)
- **API Documentation**: All public APIs now have proper go doc comments
- **Error Handling**: Consistent use of error wrapping with `%w` throughout codebase

### Fixed
- Fixed race condition that could occur when starting consumers while modifying the consumer map
- Fixed potential nil pointer issues in metrics collection when registry is nil

### Internal
- Added `registry` field to `Manager` and `hop` structs for metrics registration
- Updated `NewManager` signature to accept optional `prometheus.Registerer`
- Updated all tests to compile with new `NewManager` signature
- Added `time` import for connection duration tracking

## [0.1.0] - 2024-01-15

### Added
- Initial release of Hop RabbitMQ client
- Core features: auto-reconnect, exponential backoff, graceful shutdown
- Support for queue and exchange configuration
- Basic consumer management with `Manager`
- Structured logging with zerolog
- Example implementations in `examples/` directory
