package restfulwrapper

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tekkamanendless/httperror"
)

func TestError(t *testing.T) {
	t.Run("APIBodyError", func(t *testing.T) {
		input := fmt.Errorf("error-1")
		err := NewAPIBodyError(input)
		require.NotNil(t, err)
		assert.ErrorIs(t, err, input)
		assert.ErrorIs(t, err, httperror.ErrStatusBadRequest)
		assert.Equal(t, "error-1", err.Error())

		baseErr := &APIBodyError{}
		if assert.ErrorAs(t, err, &baseErr) {
			assert.NotNil(t, baseErr.apiResponseError)
			assert.Equal(t, input, baseErr.bodyError)
		}
	})
	t.Run("APIHeaderParameterError", func(t *testing.T) {
		input := fmt.Errorf("error-1")
		err := NewAPIHeaderParameterError("key", input)
		require.NotNil(t, err)
		assert.ErrorIs(t, err, input)
		assert.ErrorIs(t, err, httperror.ErrStatusBadRequest)
		assert.Equal(t, "error-1", err.Error())

		baseErr := &APIHeaderParameterError{}
		if assert.ErrorAs(t, err, &baseErr) {
			assert.Equal(t, "key", baseErr.parameter)
			assert.Equal(t, input, baseErr.parameterError)
		}
	})
	t.Run("APIPathParameterError", func(t *testing.T) {
		input := fmt.Errorf("error-1")
		err := NewAPIPathParameterError("key", input)
		require.NotNil(t, err)
		assert.ErrorIs(t, err, input)
		assert.ErrorIs(t, err, httperror.ErrStatusBadRequest)
		assert.Equal(t, "error-1", err.Error())

		baseErr := &APIPathParameterError{}
		if assert.ErrorAs(t, err, &baseErr) {
			assert.Equal(t, "key", baseErr.parameter)
			assert.Equal(t, input, baseErr.parameterError)
		}
	})
	t.Run("APIQueryParameterError", func(t *testing.T) {
		input := fmt.Errorf("error-1")
		err := NewAPIQueryParameterError("key", input)
		require.NotNil(t, err)
		assert.ErrorIs(t, err, input)
		assert.ErrorIs(t, err, httperror.ErrStatusBadRequest)
		assert.Equal(t, "error-1", err.Error())

		baseErr := &APIQueryParameterError{}
		if assert.ErrorAs(t, err, &baseErr) {
			assert.Equal(t, "key", baseErr.parameter)
			assert.Equal(t, input, baseErr.parameterError)
		}
	})
	t.Run("APIResponseError", func(t *testing.T) {
		t.Run("Custom Message", func(t *testing.T) {
			err := NewAPIResponseError(http.StatusForbidden, "My Message")
			require.NotNil(t, err)
			assert.ErrorIs(t, err, httperror.ErrStatusForbidden)
			assert.Equal(t, "My Message", err.Error())

			baseErr := &APIResponseError{}
			if assert.ErrorAs(t, err, &baseErr) {
				assert.Equal(t, httperror.ErrStatusForbidden, baseErr.httpError)
				assert.Equal(t, "My Message", baseErr.message)
				assert.Equal(t, http.StatusForbidden, baseErr.Code())
			}
		})
		t.Run("Default Message", func(t *testing.T) {
			err := NewAPIResponseError(http.StatusForbidden, "")
			require.NotNil(t, err)
			assert.ErrorIs(t, err, httperror.ErrStatusForbidden)
			assert.Equal(t, "Forbidden", err.Error())

			baseErr := &APIResponseError{}
			if assert.ErrorAs(t, err, &baseErr) {
				assert.Equal(t, httperror.ErrStatusForbidden, baseErr.httpError)
				assert.Equal(t, "Forbidden", baseErr.message)
				assert.Equal(t, http.StatusForbidden, baseErr.Code())
			}
		})
		t.Run("Missing Error", func(t *testing.T) {
			err := &APIResponseError{}
			require.NotNil(t, err)
			assert.Equal(t, "", err.Error())
			assert.Equal(t, http.StatusInternalServerError, err.Code())
		})
	})
}
