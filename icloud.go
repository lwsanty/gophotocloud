package main

import (
	"github.com/lwsanty/gophotocloud/engine"
	//"github.com/lwsanty/gophotocloud/photos"
	"fmt"
	"github.com/bgentry/speakeasy"
	"github.com/lwsanty/gophotocloud/drive"
)

func main() {
	login, err := speakeasy.Ask("enter your apple id: ")
	if err != nil {
		panic(err)
	}

	pass, err := speakeasy.Ask("enter your icloud pass: ")
	if err != nil {
		panic(err)
	}

	eng, err := engine.NewEngine(login, pass)
	if err != nil {
		panic(err)
	}

	fitems, cookie, token, err := drive.GetFolderItems(eng, "root")
	if err != nil {
		panic(err)
	}

	fitems_links, err := drive.GetFileItemsUrls(fitems, eng, cookie, token)
	if err != nil {
		panic(err)
	}
	fmt.Println("================================================================")
	fmt.Println("================================================================")
	fmt.Println("================================================================")
	for i := range fitems_links.Items {
		fmt.Println("name: ", fitems.Items[i].Name)
		fmt.Println("type: ", fitems.Items[i].Type)
		fmt.Println("url: ", fitems.Items[i].Url)
	}

	/*
		iclouddrive, err2 := drive.NewD(eng)
		if err2 != nil {
			panic(err2)
		}

		fmt.Println(iclouddrive.Urls)

		if err := photos.PrintContent(total); err != nil {
			panic(err)
		}

			if err := photos.DownloadContent(total); err != nil {
				panic(err)
			}
	*/
}
