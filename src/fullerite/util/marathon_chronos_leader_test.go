package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLeader(t *testing.T) {
	oldGetLeaderURL := getLeaderURL
	defer func() { getLeaderURL = oldGetLeaderURL }()

	oldHostname := hostname
	defer func() { hostname = oldHostname }()

	tests := []struct {
		ourHostname string
		rawResponse string
		expected    bool
		msg         string
	}{
		{"thequeen", "{\"leader\":\"thequeen:2017\"}", true, "Should return true when hostnames match"},
		{"thequeen", "{\"leader\":\"thequeen\"}", true, "Should return true when hostnames match and there's not port"},
		{"notthequeen", "{\"leader\":\"thequeen:2017\"}", false, "Should return false when hostnames don't match"},
		{"foobar", "", false, "Should return false on empty response"},
		{"foobar", "{\"leder\":\"me\"}", false, "Should return false when \"leader\" is not in the response"},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, test.rawResponse)
		}))
		defer ts.Close()

		getLeaderURL = func(ip string, _ string) string { return ts.URL }
		hostname = func() (string, error) { return test.ourHostname, nil }

		actual, _ := IsLeader("", "", http.Client{})

		assert.Equal(t, test.expected, actual, test.msg)
	}
}
