package metrics

// MetricsCollector define a interface para coleta de métricas.
// Implementações podem ser Prometheus, OpenTelemetry, ou qualquer outro sistema.
type MetricsCollector interface {
	// Counter retorna uma métrica do tipo Counter com as labels fornecidas.
	Counter(name string, labels ...string) Counter
	// Gauge retorna uma métrica do tipo Gauge com as labels fornecidas.
	Gauge(name string, labels ...string) Gauge
	// Histogram retorna uma métrica do tipo Histogram com as labels fornecidas.
	Histogram(name string, labels ...string) Histogram
	// Registerer retorna o registerer subjacente para registro de métricas.
	// Pode ser nil se o collector não suportar registro explícito.
	Registerer() any
}

// Counter representa uma métrica counter.
type Counter interface {
	Inc()
	Add(float64)
}

// Gauge representa uma métrica gauge.
type Gauge interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
}

// Histogram representa uma métrica histograma para medir distribuição de valores.
type Histogram interface {
	Observe(value float64)
}
