package promsum

// This program will create billing reports for a given Prometheus query for a given period and recording interval.
// Results are written to remote storage.
// It performs this by:
// - determining what gaps exist in the the billing data in storage
// - filling the gaps using data from Prometheus query:
//     - segment missing periods into blocks the size of the recording interval
//     - query prometheus for rate of range series of scalar values for period
//     - sum length of interval * rate for all time series in range
//     - write billing record with query, range, amount, and unit into storage

type Promsum interface {
	// Meter creates a billing record for a given range and Prometheus query. It does this by summing usage
	// between each Prometheus instant vector by multiplying rate against against the length of the interval.
	Meter(pqlQuery string, rng Range) (BillingRecord, error)
}
