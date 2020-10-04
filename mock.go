// Package mock contains a mock cchat backend.
package mock

import (
	"github.com/diamondburned/cchat-mock/internal/service"
	"github.com/diamondburned/cchat/services"
)

func init() {
	services.RegisterService(&service.Service{})
}
