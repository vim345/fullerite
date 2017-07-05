package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	hostname = os.Hostname

	getLeaderURL = func(host string, endpoint string) string { return fmt.Sprintf("http://%s/%s", host, endpoint) }
)

type httpError struct {
	Status int
}

type leaderError struct {
	Reason string
}

func (e httpError) Error() string {
	return fmt.Sprintf("%s: %s", http.StatusText(e.Status), e.Status)
}

func (e leaderError) Error() string {
	return e.Reason
}

// IsLeader checks if a given host is the marathon leader
func IsLeader(host string, endpoint string, client http.Client) (bool, error) {
	url := getLeaderURL(host, endpoint)

	contents, err := GetWrapper(url, client)
	if err != nil {
		return false, err
	}

	var leadermap map[string]string

	if decodeErr := json.Unmarshal(contents, &leadermap); decodeErr != nil {
		return false, decodeErr
	}

	leader, exists := leadermap["leader"]
	if !exists {
		return false, leaderError{"Could not find \"leader\" in leader JSON"}
	}

	s := strings.Split(leader, ":")

	h, err := hostname()
	if err != nil {
		return false, err
	}

	return s[0] == h, nil
}

// GetWrapper performs a get against a URL and return either the body of the response or an error
func GetWrapper(url string, client http.Client) ([]byte, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, httpError{r.StatusCode}
	}
	contents, _ := ioutil.ReadAll(r.Body)

	return []byte(contents), nil
}
