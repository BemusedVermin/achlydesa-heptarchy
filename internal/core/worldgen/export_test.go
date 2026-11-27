package worldgen

// PeakChunks is the most chunks the source's verify stage held at once, or -1 before it has run.
func PeakChunks(src *Source) int {
	for _, bound := range src.stages {
		if result, ok := bound.memo.value.(*repair); ok {
			return result.peakChunks
		}
	}
	return -1
}

// RegisterVerifyWorkers adds the verify stage with a fixed number of summary workers.
func RegisterVerifyWorkers(stages *Stages, workers int) error {
	return registerVerifyWorkers(stages, workers)
}
