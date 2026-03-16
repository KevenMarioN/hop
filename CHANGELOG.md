# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Dynamic Metrics System**: Fully decoupled metrics collection supporting multiple backends
  - New `MetricsCollector` interface for custom metric implementations
  - `PrometheusCollector` implementation for Prometheus metrics
  - `OpenTelemetryCollector` implementation for OpenTelemetry metrics
  - `MultiCollector` for using multiple backends simultaneously
  - `NopCollector` for disabled metrics (default behavior)
- **Prometheus Metrics Integration**: Added optional metrics collection with `WithMetrics()` option
  - Metrics: `hop_messages_consumed_total`, `hop_consumption_errors_total`, `hop_reconnects_total`, `hop_connection_duration_seconds`, `hop_active_consumers`, `hop_message_processing_duration_seconds`
- **OpenTelemetry Metrics Integration**: Added OpenTelemetry collector for modern observability stacks
- **ConsumerBuilder**: New fluent API for building immutable Consumer configurations via `protocol.NewConsumerBuilder()`
- **Improved Documentation**: Added comprehensive go doc comments to all public types and functions
  - Documented `protocol.Message` struct with embedded fields from amqp.Delivery
  - Documented `protocol.Handler` function signature
  - Documented `Consumer.Execute()` method with usage notes
- **HTTP Metrics Endpoint**: New `metrics.Handler()` for exposing Prometheus metrics
- **Backward Compatibility**: Added `WithPrometheusMetrics()` convenience wrapper for existing code

### Changed
- **Metrics Architecture**: Refactored metrics from global variables to interface-based system
  - `Manager` now accepts `MetricsCollector` instead of `prometheus.Registerer`
  - `WithMetrics()` now accepts `MetricsCollector` interface
  - Default behavior uses `NopCollector` when no metrics configured
- **Race Condition Fix**: Fixed potential race condition in `Manager.Start()` by copying consumer map before iteration
- **Performance Optimization**: Removed per-message debug logging from `Consumer.Execute()` (metrics provide observability)
- **API Documentation**: All public APIs now have proper go doc comments
- **Error Handling**: Consistent use of error wrapping with `%w` throughout codebase

### Fixed
- Fixed race condition that could occur when starting consumers while modifying the consumer map
- Fixed potential nil pointer issues in metrics collection when registry is nil

### Internal
- Added `collector` field to `Manager` and `hop` structs for metrics collection
- Updated `NewManager` signature to accept `MetricsCollector`
- Updated all tests to compile with new `NewManager` signature
- Added `time` import for connection duration tracking
- Created new files: `metrics/collector.go`, `metrics/nop.go`, `metrics/prometheus.go`, `metrics/multi.go`

## [0.1.0] - 2024-01-15

### Added
- Initial release of Hop RabbitMQ client
- Core features: auto-reconnect, exponential backoff, graceful shutdown
- Support for queue and exchange configuration
- Basic consumer management with `Manager`
- Structured logging with zerolog
- Example implementations in `examples/` directory
