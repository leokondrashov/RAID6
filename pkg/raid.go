package pkg

func CheckSumMatrix(d, c int) (matrix, error) {
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

// Simplified version of CheckSumMatrix that uses matrix inversion.
func CheckSumMatrixWithInv(d, c int) (matrix, error) {
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
	tranform, err := top.Invert()
	if err != nil {
		return nil, err
	}

	// apply the transformation to the whole matrix
	m, err = m.Multiply(tranform)
	if err != nil {
		return nil, err
	}

	return m, nil
}
