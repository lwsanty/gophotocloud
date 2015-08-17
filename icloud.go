package main
import (
	"github.com/mig2/icloud/engine"
	"github.com/lwsanty/gophotocloud/photos"
)
func main() {
	eng, err := engine.NewEngine("login", "pass")
	if err != nil {
		panic(err)
	}

	_, err2 := photos.NewP(eng)
	if err2 != nil {
		panic(err2)
	}
	/*
	if err := photos.PrintContent(total); err != nil {
		panic(err)
	}

	if err := photos.DownloadContent(total); err != nil {
		panic(err)
	}
	*/
}


