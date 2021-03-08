package auth

import (
	"errors"
	"net/http"

	"github.com/riposo/riposo/pkg/api"
	"go.uber.org/multierr"
)

// OneOf adds support for multi-method authentication.
func OneOf(multiple ...Method) Method {
	return oneOf(multiple)
}

type oneOf []Method

// Authenticate implements Method interface.
func (mm oneOf) Authenticate(r *http.Request) (*api.User, error) {
	var unauthErr error
	for _, sub := range mm {
		user, err := sub.Authenticate(r)
		if errors.Is(err, ErrUnauthenticated) {
			unauthErr = err
			continue
		} else if err != nil {
			return nil, err
		} else if user != nil {
			return user, nil
		}
	}

	if unauthErr != nil {
		return nil, unauthErr
	}
	return nil, Errorf("no authentication methods enabled")
}

// Close implements io.Closer interface.
func (mm oneOf) Close() (err error) {
	for _, sub := range mm {
		err = multierr.Append(err, sub.Close())
	}
	return
}
