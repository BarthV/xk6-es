package es

import (
	xk6es "github.com/barthv/xk6-es/internal"
	"go.k6.io/k6/output"
)

func init() {
	output.RegisterExtension("xk6-es", func(p output.Params) (output.Output, error) {
		return xk6es.New(p)
	})
}
