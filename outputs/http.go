package outputs

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

func postReport(json []byte, u *url.URL) error {
	resp, err := http.Post(
		u.String(),
		"application/json",
		bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	fmt.Printf("status code from report URL %s: %s\n", resp.Request.URL.String(), resp.Status)

	return nil
}
