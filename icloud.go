package main

import (
	"github.com/lwsanty/gophotocloud/engine"
	//"github.com/lwsanty/gophotocloud/photos"
	"github.com/lwsanty/gophotocloud/drive"
	"fmt"
)

func main() {
	eng, err := engine.NewEngine("title", "cypher_")
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
		if fitems_links.Items[i].Type == "FOLDER" {
			fitemz, _, _, err := drive.GetFolderItems(eng, fitems_links.Items[i].Id)
			if err != nil {
				panic(err)
			}

			fitemzz, err := drive.GetFileItemsUrls(fitemz, eng, cookie, token)
			if err != nil {
				panic(err)
			}

			fmt.Println("name: ", fitemzz.Items[i].Name)
			fmt.Println("type: ", fitemzz.Items[i].Type)
			fmt.Println("url: ", fitemzz.Items[i].Url)
			break
		}
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
