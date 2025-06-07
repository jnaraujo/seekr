package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/jnaraujo/seekr/cmd"
)

func main() {
	// cleanup := profileApp()
	// defer cleanup()
	cmd.Execute()
}

func profileApp() func() {
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}

	return func() {
		pprof.StopCPUProfile()
		cpuFile.Close()
	}
}
