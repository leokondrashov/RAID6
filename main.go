package main

import (
	"flag"
	"fmt"

	"github.com/lkondras/RAID6/pkg"
)

var (
	dataDiskCount   = flag.Int("data", 6, "Number of data disks")
	parityDiskCount = flag.Int("parity", 2, "Number of parity disks")
	classicRAID6    = flag.Bool("classic", false, "Use classic RAID6 Linux implementation")
	directory       = flag.String("dir", "data", "Directory to use for the shards")
)

func main() {

	flag.Parse()

	if *classicRAID6 && (*dataDiskCount != 6 || *parityDiskCount != 2) {
		fmt.Println("Classic RAID6 requires 6 data disks and 2 parity disks")
		return
	}

	var m pkg.Matrix
	if *classicRAID6 {
		m, _ = pkg.CheckSumMatrixClassic()
	} else {
		m, _ = pkg.CheckSumMatrix(*dataDiskCount, *parityDiskCount)
	}

	operation := flag.CommandLine.Arg(0)
	if operation == "store" {
		file := flag.CommandLine.Arg(1)
		fmt.Println("Storing file", file)
		err := pkg.StoreFile(file, m, *directory)
		if err != nil {
			fmt.Println("Error storing file:", err)
		}
	} else if operation == "recover" {
		fmt.Println("Recovering data")
		pkg.RecoverData(m, *directory)
	} else if operation == "read" {
		file := flag.CommandLine.Arg(1)
		fmt.Println("Reading to file", file)
		pkg.ReadFile(file, m, *directory)
	} else {
		fmt.Println("Invalid operation")
		return
	}

}
