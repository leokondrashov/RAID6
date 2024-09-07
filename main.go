package main

import (
	"fmt"

	"github.com/lkondras/RAID6/pkg"
)

func main() {
	m, _ := pkg.CheckSumMatrix(4, 4)
	fmt.Println(m)
}
