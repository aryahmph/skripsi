// Package ucase
package ucase

import (
	"en-ticket/internal/appctx"
	"en-ticket/internal/consts"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHealthCheck_Serve(t *testing.T) {
	svc := NewHealthCheck()

	t.Run("test health check", func(t *testing.T) {
		req := &appctx.Data{}
		result := svc.Serve(req)

		assert.Equal(t, appctx.Response{
			Name:    consts.ResponseSuccess,
			Message: "ok",
		}, result)
	})
}
