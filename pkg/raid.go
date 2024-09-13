package pkg

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
