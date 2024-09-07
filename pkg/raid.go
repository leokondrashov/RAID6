package pkg

func CheckSumMatrix(d, c int) (matrix, error) {
	m, err := newMatrix(d+c, d)
	if err != nil {
		return nil, err
	}
	for i := 0; i < d; i++ {
		m[i][i] = 1
	}
	for i := 0; i < c; i++ {
		for j := 0; j < d; j++ {
			m[d+i][j] = galExp(byte(j+1), i)
		}
	}
	return m, nil
}
