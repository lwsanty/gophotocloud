package reminders

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/kr/pretty"
)

var app *ICloudRemindersApp

func TestReminder(t *testing.T) {
	var body []byte
	var e error
	if body, e = ioutil.ReadFile("icloud.reminders.response"); e != nil {
		t.Fatal(e)
	}
	app = &ICloudRemindersApp{}
	if e = json.Unmarshal(body, app); e != nil {
		t.Fatal(e)
	}

	fmt.Printf("%# v", pretty.Formatter(app))
	//t.Errorf("Reverse(%q) == %q, want %q", c.in, got, c.want)
}

func TestParent(t *testing.T) {
	a, _ := app.NewReminder("test", "lo")
	fmt.Printf("%# v", pretty.Formatter(a))

	app.sync()
}
