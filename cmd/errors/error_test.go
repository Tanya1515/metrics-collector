package retryerr

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestCheckErrorType(t *testing.T) {
	client := resty.New().SetTimeout(5 * time.Millisecond)

	t.Run("Test:", func(t *testing.T) {
		_, err := client.R().Get("http://0.0.0.0:8080/ping")

		resultErr := CheckErrorType(err)
		assert.Equal(t, true, resultErr)
	})
}
