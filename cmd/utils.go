package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func waitForHttpResponse(done chan bool) {
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

func checkIfFileExists(fileName string) (bool, error) {
	_, errFileExist := os.Stat(fileName)
	if errFileExist != nil && errors.Is(errFileExist, os.ErrNotExist) {
		return false, nil
	}

	if errFileExist != nil {
		return false, errFileExist
	}

	return true, nil
}
