package ldmigration

import "fmt"

// ExecutionOrder represents the various execution modes this SDK can operate
// under while performing migration-assisted reads.
type ExecutionOrder uint8

const (
	// Serial execution ensures the authoritative read will always complete execution before executing the
	// non-authoritative read.
	Serial ExecutionOrder = iota
	// Random execution randomly decides if the authoritative read should execute first or second.
	Random
	// Concurrent executes both reads in separate go routines, and waits until both calls have finished before
	// proceeding.
	Concurrent
)

// MigrationOp represents a type of migration operation; namely, read or write.
type MigrationOp uint8

const (
	// Read denotes a read-related migration operation.
	Read MigrationOp = iota
	// Write denotes a write-related migration operation.
	Write
)

func (o MigrationOp) String() string {
	switch o {
	case Read:
		return "read"
	case Write:
		return "write"
	default:
		return fmt.Sprintf("%d", int(o))
	}
}

// ConsistencyCheck records the results of a consistency check and the ratio at
// which the check was sampled.
//
// For example, a sampling ratio of 10 indicts this consistency check was
// sampled approximately once every ten operations.
type ConsistencyCheck struct {
	consistent    bool
	samplingRatio int
}

// NewConsistencyCheck creates a new consistency check reflecting the provided values.
func NewConsistencyCheck(wasConsistent bool, samplingRatio int) *ConsistencyCheck {
	return &ConsistencyCheck{
		consistent:    wasConsistent,
		samplingRatio: samplingRatio,
	}
}

// Consistent returns whether or not the check returned a consistent result.
func (c ConsistencyCheck) Consistent() bool {
	return c.consistent
}

// SamplingRatio returns the 1 in x sampling ratio used to determine if the consistency check should be run.
func (c ConsistencyCheck) SamplingRatio() int {
	return c.samplingRatio
}

// MigrationOrigin represents the source of origin for a migration-related operation.
type MigrationOrigin int

const (
	// Old represents the technology source we are migrating away from.
	Old MigrationOrigin = iota
	// New represents the technology source we are migrating towards.
	New
)

func (o MigrationOrigin) String() string {
	switch o {
	case Old:
		return "old"
	case New:
		return "new"
	default:
		return fmt.Sprintf("%d", int(o))
	}
}

// MigrationStage denotes one of six possible stages a technology migration could be a part of.
type MigrationStage int

const (
	// Off Stage 1 - migration hasn't started, "old" is authoritative for reads and writes
	Off MigrationStage = iota

	// DualWrite Stage 2 - write to both "old" and "new", "old" is authoritative for reads
	DualWrite

	// Shadow Stage 3 - both "new" and "old" versions run with a preference for "old"
	Shadow

	// Live Stage 4 - both "new" and "old" versions run with a preference for "new"
	Live

	// RampDown Stage 5 only read from "new", write to "old" and "new"
	RampDown

	// Complete Stage 6 - migration is done
	Complete
)

// String converts a MigrationStage into its string representation.
func (s MigrationStage) String() string {
	switch s {
	case Off:
		return "off" //nolint:goconst
	case DualWrite:
		return "dualwrite"
	case Shadow:
		return "shadow"
	case Live:
		return "live"
	case RampDown:
		return "rampdown"
	case Complete:
		return "complete"
	default:
		return "off"
	}
}

// NewMigrationStageFromString is a convenience method for creating a migration stage enum from a simple string.
func NewMigrationStageFromString(val string) (MigrationStage, error) {
	switch val {
	case "off":
		return Off, nil
	case "dualwrite":
		return DualWrite, nil
	case "shadow":
		return Shadow, nil
	case "live":
		return Live, nil
	case "rampdown":
		return RampDown, nil
	case "complete":
		return Complete, nil
	default:
		return Off, fmt.Errorf("invalid stage %s provided", val)
	}
}
