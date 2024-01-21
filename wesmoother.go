// Copyright 2024 Kurt Grutzmacher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package smoother implements the Whittaker-Eilers Smoothing function for a slice of float64 values
//
// The function is based on the work by Paul H.C. Eilers "A Perfect Smoother".
// The paper and supporting info can be found at https://pubs.acs.org/doi/full/10.1021/ac034173t
package smoother

import (
	"errors"

	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
)

// speye creates an identity matrix of size N in Compressed Sparse Row (CSR) format.
// First we initialize three slices: data, indices, and indptr, which are used to create the CSR matrix.
// The data slice is then filled with 1s, and the indices and indptr slices are filled with increasing
// integers from 0 to N-1.
func speye(n int) *sparse.CSR {
	data := make([]float64, n)
	indices := make([]int, n)
	indptr := make([]int, n+1)

	for i := 0; i < n; i++ {
		data[i] = 1
		indices[i] = i
		indptr[i] = i
	}
	indptr[n] = n

	return sparse.NewCSR(n, n, indices, indptr, data)
}

// vecDiff calculates the element-wise difference between two slices a and b, which should be the same length.
// A new slice where each element is the difference between the corresponding elements in a and b is returned.
func vecDiff(a, b []float64) []float64 {
	diff := make([]float64, len(a))
	for i := range a {
		diff[i] = a[i] - b[i]
	}
	return diff
}

// differenceMatrix creates a difference matrix of size n with order d by first creating a vector of coefficients,
// which are then used to fill a Compressed Spares Row (CSR) matrix.
func differenceMatrix(n int, order int) *sparse.CSR {
	coeffs := make([]float64, 2*order+1)
	coeffs[order] = 1.0

	for i := 0; i < order; i++ {
		coeffs = vecDiff(coeffs[:len(coeffs)-1], coeffs[1:])
	}

	nRows := n - order
	data := make([]float64, nRows*n)
	indices := make([]int, nRows*n)
	indptr := make([]int, nRows+1)

	idx := 0
	for i := 0; i < nRows; i++ {
		indptr[i] = idx
		for j := 0; j <= order; j++ {
			data[idx] = coeffs[j]
			indices[idx] = i + j
			idx++
		}
	}
	indptr[nRows] = idx

	return sparse.NewCSR(nRows, n, indptr, indices, data)
}

// WESmoother applies the Whittaker-Eilers smoothing function to a given data series y with a specified
// parameter lambda and order d. It returns the smoothed series and an error if the Cholesky decomposition fails.
// The data series is assumed to be collected from an equal sample rate.
//
// The function is based on the work by Paul H.C. Eilers "A Perfect Smoother".
// A larger lambda will increase the smoothness of the series, but may also result in a loss of detail.
func WESmoother(y []float64, lambda float64, d int) ([]float64, error) {
	m := len(y)

	E := speye(m)
	D := differenceMatrix(m, d)

	// Convert sparse matrices to dense
	EDense := mat.DenseCopyOf(E.ToDense())
	DDense := mat.DenseCopyOf(D.ToDense())

	// Compute D' * D
	DTD := &mat.Dense{}
	DTD.Mul(DDense.T(), DDense)

	// Scale D' * D by lambda
	DTD.Scale(lambda, DTD)

	// Add E and lambda * D' * D
	A := &mat.Dense{}
	A.Add(EDense, DTD)
	r, _ := A.Dims()
	sym := mat.NewSymDense(r, nil)

	// Copy the upper triangular part of A to sym
	for i := 0; i < r; i++ {
		for j := i; j < r; j++ {
			sym.SetSym(i, j, A.At(i, j))
		}
	}

	// Convert A to a dense matrix and compute its Cholesky decomposition
	var chol mat.Cholesky
	ok := chol.Factorize(sym)
	if !ok {
		return nil, errors.New("cholesky decomposition failed")
	}

	// Solve the system of linear equations C * z = y for z
	z := make([]float64, m)
	b := mat.NewVecDense(m, y)
	x := mat.NewVecDense(m, z)
	err := chol.SolveVecTo(x, b)

	return z, err
}
