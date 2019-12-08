package uof

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	inner := fmt.Errorf("get failed")
	ae := ApiError{URL: "url", Inner: inner}
	assert.Equal(t, inner, ae.Unwrap())

	e := E("api", ae)
	var err error
	err = e

	var s string
	var ae2 ApiError
	if errors.As(err, &ae2) {
		s = ae2.Error()
	}
	assert.Equal(t, "uof api error url: url, inner: get failed", s)

	var e2 Error
	if errors.As(err, &e2) {
		s = e2.Error()
	}
	assert.Equal(t, "uof error op: api, inner: uof api error url: url, inner: get failed", s)
}

func TestInnerError(t *testing.T) {
	inner := fmt.Errorf("some inner error")
	ue := Notice("operation", inner)
	err := ue.Unwrap()
	assert.Equal(t, inner, err)
}
