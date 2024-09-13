package pkg

import (
	"fmt"
	"os"
	"strconv"
)

const (
	lengthSize = strconv.IntSize / 8
)

// General case of checksum matrix.
// Has the property that the first d rows are identity matrix
// and it is invertible if C rows are removed.
func CheckSumMatrix(d, c int) (Matrix, error) {
	m, err := vandermonde(d+c, d)
	if err != nil {
		return nil, err
	}

	// gaussian elimination to make the first d rows identity matrix
	for i := 0; i < d; i++ {
		if m[i][i] == 0 {
			for j := i + 1; j < d; j++ {
				if m[i][j] == 0 {
					continue
				}

				// swap columns i and j
				for k := 0; k < d+c; k++ {
					m[k][i], m[k][j] = m[k][j], m[k][i]
				}
				break
			}
		}

		f := m[i][i]
		for k := 0; k < d+c; k++ {
			m[k][i] = galDivide(m[k][i], f)
		}

		for j := 0; j < d; j++ {
			if (i == j) || (m[i][j] == 0) {
				continue
			}

			f := m[i][j]
			for k := 0; k < d+c; k++ {
				m[k][j] = galAdd(m[k][j], galMultiply(m[k][i], f)) // m[j] = m[j] - m[i] * f
			}
		}
	}

	return m, nil
}

// General case of checksum matrix.
// Has the property that the first d rows are identity matrix
// and it is invertible if C rows are removed.
// Simplified implementation that uses matrix inversion.
func CheckSumMatrixWithInv(d, c int) (Matrix, error) {
	m, err := vandermonde(d+c, d)
	if err != nil {
		return nil, err
	}

	// top d rows of the matrix
	top, err := m.SubMatrix(0, 0, d, d)
	if err != nil {
		return nil, err
	}

	// invert the top d rows, the result would correspond to column manipulation during gaussian elimination
	transform, err := top.Invert()
	if err != nil {
		return nil, err
	}

	// apply the transformation to the whole matrix
	m, err = m.Multiply(transform)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Linux RAID6 classic checksum matrix.
// This is a special case of CheckSumMatrix with d=6 and c=2.
// The matrix is:
// 1  0  0  0  0  0
// 0  1  0  0  0  0
// 0  0  1  0  0  0
// 0  0  0  1  0  0
// 0  0  0  0  1  0
// 0  0  0  0  0  1
// 1  1  1  1  1  1
// 32 16 8  4  2  1
func CheckSumMatrixClassic() (Matrix, error) {
	d := 6
	c := 2

	m, err := newMatrix(d+c, d)
	if err != nil {
		return nil, err
	}

	// top d rows are identity matrix
	for i := 0; i < d; i++ {
		m[i][i] = 1
	}

	// row of ones
	for j := 0; j < d; j++ {
		m[d][j] = 1
	}

	// row of powers of 2
	for j := 0; j < d; j++ {
		m[d+1][j] = galExp(2, 5-j)
	}

	return m, nil
}

func (m Matrix) MultiplyData(data []byte) ([][]byte, error) {
	d := len(m[0])

	// Split the data into d-sized chunks
	chunks := make([][]byte, 0)
	chunkLen := len(data) / d
	for i := 0; i < len(data); i += chunkLen {
		chunks = append(chunks, data[i:i+chunkLen])
	}

	// Multiply the chunks by the matrix
	shards, err := m.Multiply(chunks)
	if err != nil {
		return nil, err
	}

	return shards, nil
}

// Stores a file of arbitrary size in data shards using the provided matrix.
// First 8 bytes of the file are used to store the file size.
func StoreFile(file string, m Matrix, directory string) error {
	// Create directory if it does not exist
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	// Read the file
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Prepend the file size to the data
	length := len(data)
	lengthArray := make([]byte, lengthSize)
	for i := 0; i < lengthSize; i++ {
		lengthArray[i] = byte(length >> (8 * i))
	}
	data = append(lengthArray, data...)
	length += strconv.IntSize / 8

	// Append padding if necessary
	paddings := 0
	if length%len(m[0]) != 0 {
		paddings = len(m[0]) - length%len(m[0])
	}
	data = append(data, make([]byte, paddings)...)

	// Split the data into shards
	// Also calculates the parity shards
	shards, err := m.MultiplyData(data)
	if err != nil {
		return err
	}

	// Write the shards to the directory
	for i, shard := range shards {
		err = os.WriteFile(fmt.Sprintf("%s/shard%d", directory, i), shard, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadFile(file string, m Matrix, directory string) error {
	d := len(m[0])
	c := len(m) - d

	// Create directory if it does not exist
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist")
	}

	// Read the shards
	shards := make([][]byte, d+c)
	for i := 0; i < d+c; i++ {
		shard, err := os.ReadFile(fmt.Sprintf("%s/shard%d", directory, i))
		if err != nil {
			return fmt.Errorf("error reading shard %d, consider running recovery", i)
		}
		shards[i] = shard
	}

	// Check the shards
	parity := shards[d:]
	data := shards[:d]
	restored, err := m[d:].Multiply(data)
	if err != nil {
		return fmt.Errorf("error checking parity: %w", err)
	}

	// Compare the parity with the restored data
	// Single error means that parity disk has the error -> recoverable
	// Multiple errors mean that data disks have the error -> unrecoverable
	parityErrors := 0
	for i := 0; i < c; i++ {
		for j := 0; j < len(restored[i]); j++ {
			if restored[i][j] != parity[i][j] {
				parityErrors++
			}
		}
	}

	if parityErrors > 1 {
		return fmt.Errorf("too many parity errors, unrecoverable")
	} else if parityErrors == 1 {
		fmt.Println("parity has error, consider running recovery")
	}

	// Write the data to the file
	rawData := make([]byte, 0)
	for i := 0; i < len(data); i++ {
		rawData = append(rawData, data[i]...)
	}
	length := 0
	for i := 0; i < lengthSize; i++ {
		length |= int(rawData[i]) << (8 * i)
	}

	err = os.WriteFile(file, rawData[lengthSize:length+lengthSize], 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func RecoverData(m Matrix, directory string) error {
	d := len(m[0])

	// Create directory if it does not exist
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist")
	}

	// Read the shards
	shards := make([][]byte, 0)
	presentShards := make([]int, 0)
	missingShards := make([]int, 0)
	for i := 0; i < len(m); i++ {
		shard, err := os.ReadFile(fmt.Sprintf("%s/shard%d", directory, i))
		if err == nil {
			presentShards = append(presentShards, i)
			shards = append(shards, shard)
		} else {
			missingShards = append(missingShards, i)
		}
	}

	if len(presentShards) < d {
		return fmt.Errorf("too many missing shards, unrecoverable")
	} else if len(missingShards) == 0 {
		return nil
	}

	// Calculate the missing data shards
	recoveryRows := make([][]byte, d)
	for i := 0; i < d; i++ {
		recoveryRows[i] = m[presentShards[i]]
	}
	tmp, err := newMatrixData(recoveryRows)
	if err != nil {
		return fmt.Errorf("error creating recovery matrix: %w", err)
	}
	recoveryMatrix, err := tmp.Invert()
	if err != nil {
		return fmt.Errorf("error inverting recovery matrix: %w", err)
	}

	recoveredData, err := recoveryMatrix.Multiply(shards[:d])
	if err != nil {
		return fmt.Errorf("error recovering data: %w", err)
	}

	// Write the recovered data to the directory
	for i, shard := range recoveredData {
		err = os.WriteFile(fmt.Sprintf("%s/shard%d", directory, i), shard, 0644)
		if err != nil {
			return fmt.Errorf("error writing recovered shard %d: %w", i, err)
		}
	}

	// Recompute the parity shards
	parity, err := m[d:].Multiply(recoveredData)
	if err != nil {
		return fmt.Errorf("error computing parity: %w", err)
	}
	for i, shard := range parity {
		err = os.WriteFile(fmt.Sprintf("%s/shard%d", directory, d+i), shard, 0644)
		if err != nil {
			return fmt.Errorf("error writing parity shard %d: %w", d+i, err)
		}
	}

	return nil
}
