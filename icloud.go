package main

import (
	"github.com/lwsanty/gophotocloud/engine"
	//"github.com/lwsanty/gophotocloud/photos"
	"github.com/lwsanty/gophotocloud/drive"
	"fmt"
)

func main() {
	eng, err := engine.NewEngine("login", "password")
	if err != nil {
		panic(err)
	}

	iclouddrive, err2 := drive.NewD(eng)
	if err2 != nil {
		panic(err2)
	}

	fmt.Println(iclouddrive.Urls)
	/*
		if err := photos.PrintContent(total); err != nil {
			panic(err)
		}

		if err := photos.DownloadContent(total); err != nil {
			panic(err)
		}
	*/
}
