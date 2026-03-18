package metrics

type nopCounter struct{}

func (nc *nopCounter) Inc()          {}
func (nc *nopCounter) Add(v float64) {}

type nopGauge struct{}

func (ng *nopGauge) Set(v float64) {}
func (ng *nopGauge) Inc()          {}
func (ng *nopGauge) Dec()          {}
func (ng *nopGauge) Add(v float64) {}

type nopHistogram struct{}

func (nh *nopHistogram) Observe(value float64) {}

// NopCollector is a collector that does nothing.
// Useful when metrics are disabled.
var NopCollector = &nopCollector{}

type nopCollector struct{}

func (nc *nopCollector) Counter(name string, labels ...string) Counter {
	return &nopCounter{}
}

func (nc *nopCollector) Gauge(name string, labels ...string) Gauge {
	return &nopGauge{}
}

func (nc *nopCollector) Histogram(name string, labels ...string) Histogram {
	return &nopHistogram{}
}

func (nc *nopCollector) Registerer() any {
	return nil
}
