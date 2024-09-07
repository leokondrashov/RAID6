package main

import (
	"fmt"

	"github.com/lkondras/RAID6/pkg"
)

func main() {
	m, _ := pkg.CheckSumMatrix(3, 3)
	fmt.Println(m)

	m, _ = pkg.CheckSumMatrixWithInv(3, 3)
	fmt.Println(m)
}
