package uof

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	ae := ApiError{URL: "url", Inner: fmt.Errorf("get failed")}
	e := E("api", ae)
	var err error
	err = e
	//assert.True(t, errors.Is(e, ApiError))

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
