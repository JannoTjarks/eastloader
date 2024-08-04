package visiolink

import (
	"fmt"
	"time"
)

func WaitForHttpResponse(done chan bool) {
out:
	for {
		select {
		case <-done:
			fmt.Println("\nDownload is finished!")
			break out
		default:
			fmt.Print(".")
			time.Sleep(2 * time.Second)
		}
	}
}
