package metrics

// Metrics is a simple in-memory metrics holder.
// You can later replace this with a real Prometheus client.
type Metrics struct {
	RequestsTotal uint64
	ErrorsTotal   uint64
}

// New creates a new Metrics instance.
func New() *Metrics {
	return &Metrics{}
}

// IncRequests increments the total request counter.
func (m *Metrics) IncRequests() {
	m.RequestsTotal++
}

// IncErrors increments the total error counter.
func (m *Metrics) IncErrors() {
	m.ErrorsTotal++
}
