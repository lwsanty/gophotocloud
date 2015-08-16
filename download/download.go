package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFromUrl(url string, fileName string) {
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	if _, err := os.Stat("images/" + fileName); err == nil {
		fmt.Printf("file exists; removing old one...")
		os.Remove("images/" + fileName)
	}

	output, err := os.Create("images/" + fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}
