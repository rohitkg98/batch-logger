package syncer

type ProcessOrLog struct {
	process bool
	entry   Log
}

// Supplements Syncer with channels for concurrency
type Payloads struct {
	BatchSize int
	Stream    chan ProcessOrLog
}

// Add a log entry to payloads
func (p Payloads) Add(log Log) {
	p.Stream <- ProcessOrLog{
		process: false,
		entry:   log,
	}
}

// Manually trigger syncing
func (p Payloads) Process() {
	p.Stream <- ProcessOrLog{
		process: true,
	}
}
