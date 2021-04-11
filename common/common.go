package common

import (
	"fmt"
	"net/http"
	"strings"
)

func DefaultRequest(w http.ResponseWriter, r *http.Request) {
	availableEndpoints := []string{"/password-reset-request", "/verify-email-request"}
	fmt.Fprint(w, "List of available endpoints:\n", strings.Join(availableEndpoints[:], "\n"), "\n")
}
