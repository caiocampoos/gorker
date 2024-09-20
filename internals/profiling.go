package internals

import (
	"fmt"
	"runtime"
)

func PrintMemoryUsage(phase string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("%s: Memory Usage\n", phase)
	fmt.Printf("Alloc = %v MiB\n", bToMb(m.Alloc))
	fmt.Printf("TotalAlloc = %v MiB\n", bToMb(m.TotalAlloc))
	fmt.Printf("Sys = %v MiB\n", bToMb(m.Sys))
	fmt.Printf("NumGC = %v\n\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
