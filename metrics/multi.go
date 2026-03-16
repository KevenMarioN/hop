package metrics

// MultiCollector permite usar múltiplos collectors simultaneamente.
type MultiCollector struct {
	collectors []MetricsCollector
}

// NewMultiCollector cria um novo collector que envia métricas para múltiplos backends.
func NewMultiCollector(collectors ...MetricsCollector) *MultiCollector {
	return &MultiCollector{collectors: collectors}
}

func (mc *MultiCollector) Counter(name string, labels ...string) Counter {
	return &multiCounter{
		collectors: mc.collectors,
		name:       name,
		labels:     labels,
	}
}

func (mc *MultiCollector) Gauge(name string, labels ...string) Gauge {
	return &multiGauge{
		collectors: mc.collectors,
		name:       name,
		labels:     labels,
	}
}

func (mc *MultiCollector) Histogram(name string, labels ...string) Histogram {
	return &multiHistogram{
		collectors: mc.collectors,
		name:       name,
		labels:     labels,
	}
}

func (mc *MultiCollector) Registerer() any {
	// Retorna o primeiro registerer disponível (para compatibilidade)
	for _, c := range mc.collectors {
		if r := c.Registerer(); r != nil {
			return r
		}
	}
	return nil
}

type multiCounter struct {
	collectors []MetricsCollector
	name       string
	labels     []string
}

func (mc *multiCounter) Inc() {
	for _, c := range mc.collectors {
		c.Counter(mc.name, mc.labels...).Inc()
	}
}

func (mc *multiCounter) Add(v float64) {
	for _, c := range mc.collectors {
		c.Counter(mc.name, mc.labels...).Add(v)
	}
}

type multiGauge struct {
	collectors []MetricsCollector
	name       string
	labels     []string
}

func (mg *multiGauge) Set(v float64) {
	for _, c := range mg.collectors {
		c.Gauge(mg.name, mg.labels...).Set(v)
	}
}

func (mg *multiGauge) Inc() {
	for _, c := range mg.collectors {
		c.Gauge(mg.name, mg.labels...).Inc()
	}
}

func (mg *multiGauge) Dec() {
	for _, c := range mg.collectors {
		c.Gauge(mg.name, mg.labels...).Dec()
	}
}

func (mg *multiGauge) Add(v float64) {
	for _, c := range mg.collectors {
		c.Gauge(mg.name, mg.labels...).Add(v)
	}
}

type multiHistogram struct {
	collectors []MetricsCollector
	name       string
	labels     []string
}

func (mh *multiHistogram) Observe(value float64) {
	for _, c := range mh.collectors {
		c.Histogram(mh.name, mh.labels...).Observe(value)
	}
}
