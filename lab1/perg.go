package main

import (
	"fmt"
	"lab1/perfecthashing"
	"os"
	"runtime/pprof"
)

func main() {
	f, _ := os.Create("cpu_profile.prof")
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	keys := []string{"apple", "banana", "cherry", "date", "fig"}
	primary := perfecthashing.NewPrimaryHashTable(keys, 5)

	fmt.Println(primary.Contains("1"))
	fmt.Println(primary.Contains("apple"))

	memFile, _ := os.Create("mem_profile.prof")
	defer memFile.Close()
	pprof.WriteHeapProfile(memFile)
}
