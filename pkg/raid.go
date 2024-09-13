package pkg

import (
	"fmt"
	"os"
	"strconv"
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
	lengthArray := make([]byte, strconv.IntSize/8)
	for i := 0; i < strconv.IntSize/8; i++ {
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
	return fmt.Errorf("not implemented")
}

func RecoverData(m Matrix, directory string) error {
	return fmt.Errorf("not implemented")
}
