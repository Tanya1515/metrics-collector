package retryerr

import (
	"errors"
	"net"
	"syscall"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

func CheckErrorType(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		if (err.Code == pgerrcode.ConnectionException) || (err.Code == pgerrcode.ConnectionDoesNotExist) || (err.Code == pgerrcode.ConnectionFailure) || (err.Code == pgerrcode.InvalidTransactionInitiation) {
			return true
		}
	}

	if net, ok := err.(net.Error); ok {
		if errors.Is(net, syscall.ECONNREFUSED) || errors.Is(net, syscall.ETIMEDOUT) || errors.Is(net, syscall.EADDRNOTAVAIL) || errors.Is(net, syscall.EHOSTUNREACH) {
			return true
		}
	}

	return false
}
