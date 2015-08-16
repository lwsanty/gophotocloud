package main
import (
	"github.com/mig2/icloud/engine"
	"github.com/icloud/photos"
)
func main() {
	eng, err := engine.NewEngine("", "")
	if err != nil {
		panic(err)
	}

	err2 := photos.NewP(eng)
	if err2 != nil {
		panic(err2)
	}

}


