package loki

import (
	"fmt"
	"net/http"
	"venomouswolf/minlog/app/log"
)

func SendReloadReqeust() bool {
	l := log.GetLogger()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/-/reload", lokiPort))

	if err != nil {
		l.Error(err.Error())
		return false
	}

	if resp.StatusCode != http.StatusOK {
		l.Error(fmt.Sprintf("Error: HTTP StatusCode %d", resp.StatusCode))

		fmt.Println("Error reading response:", err)
		return false
	}
	return true
}
