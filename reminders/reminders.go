package reminders //github.com/mig2/icloud/reminders

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/mig2/icloud/engine"

	"code.google.com/p/go-uuid/uuid"
)

type ICloudRemindersApp struct {
	GUID2Name   map[string]string
	Name2GUID   map[string]string
	Engine      *engine.ICloudEngine
	Collections []ICloudReminderFolder `json:"Collections"`
	InboxItem   []interface{}          `json:"InboxItem"`
	Reminders   []ICloudReminder       `json:"Reminders"`
}

type ICloudReminderFolder struct {
	CollectionShareType interface{} `json:"collectionShareType"`
	Color               string      `json:"color"`
	CompletedCount      int         `json:"completedCount"`
	CreatedDate         []int       `json:"createdDate"`
	CreatedDateExtended interface{} `json:"createdDateExtended"`
	Ctag                string      `json:"ctag"`
	EmailNotification   interface{} `json:"emailNotification"`
	Enabled             bool        `json:"enabled"`
	GUID                string      `json:"guid"`
	IsFamily            bool        `json:"isFamily"`
	Order               int64       `json:"order"`
	Participants        interface{} `json:"participants"`
	SymbolicColor       string      `json:"symbolicColor"`
	Title               string      `json:"title"`
}

type ICloudAlarm struct {
	Description     string      `json:"description"`
	IsLocationBased bool        `json:"isLocationBased"`
	Proximity       interface{} `json:"proximity"`
	MessageType     string      `json:"messageType"`
	OnDate          []int       `json:"onDate"`
	Measurement     interface{} `json:"measurement"`
	GUID            string      `json:"guid"`
}

type ICloudReminder struct {
	Alarms              []ICloudAlarm `json:"alarms"`
	CompletedDate       interface{}   `json:"completedDate"`
	CreatedDate         []int         `json:"createdDate"`
	CreatedDateExtended int64         `json:"createdDateExtended"`
	Description         string        `json:"description"`
	DueDate             []int         `json:"dueDate"`
	DueDateIsAllDay     bool          `json:"dueDateIsAllDay"`
	Etag                string        `json:"etag"`
	GUID                string        `json:"guid"`
	IsFamily            string        `json:"isFamily"`
	LastModifiedDate    []int         `json:"lastModifiedDate"`
	Order               int64         `json:"order"`
	PGUID               string        `json:"pGuid"`
	Priority            int           `json:"priority"`
	Recurrence          interface{}   `json:"recurrence"`
	StartDate           []int         `json:"startDate"`
	StartDateIsAllDay   bool          `json:"startDateIsAllDay"`
	StartDateTz         string        `json:"startDateTz"`
	Title               string        `json:"title"`
}

type ChangeSet struct {
	updates struct {
		Reminders   []ICloudReminder       `json:"Reminders"`
		Collections []ICloudReminderFolder `json:"Collections"`
	} `json:"updates"`
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func makeStartDate() []int {
	rval := make([]int, 7)
	t := time.Now()
	rval[0] = t.Year()*10*10*10*10 + int(t.Month())*10*10 + t.Day()
	rval[1] = t.Year()
	rval[2] = int(t.Month())
	rval[3] = t.Day()
	rval[4] = t.Hour()
	rval[5] = t.Minute()
	rval[6] = t.Second()
	return rval
}

func NewApp(cloud *engine.ICloudEngine) (*ICloudRemindersApp, error) {

	var svc engine.ICloudService
	var e error
	var ok bool
	if cloud.Client == nil {
		return nil, Error("Client not logged in")
	}
	if svc, ok = cloud.Webservices["reminders"]; !ok {
		return nil, Error("No Reminders app")
	}
	if svc.Status != "active" {
		return nil, Error("Reminders service not active")
	}

	host, _, _ := net.SplitHostPort(svc.Url)
	var req *http.Request
	if req, e = http.NewRequest("GET", host+"/rd/startup", nil); e != nil {
		return nil, e
	}

	v := url.Values{}
	v.Add("clientBuildNumber", cloud.ReportedVersion.BuildNumber)
	v.Add("clientID", cloud.ClientID)
	v.Add("clientVersion", fmt.Sprintf("%.1f", float32(cloud.Version)))
	v.Add("dsid", cloud.User.Dsid)
	v.Add("lang", cloud.User.LanguageCode)
	v.Add("usertz", "US/Pacific")
	req.URL.RawQuery = v.Encode()
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Add("Origin", "https://www.icloud.com")

	var resp *http.Response
	if resp, e = cloud.Client.Do(req); e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(resp.Body); e != nil {
		return nil, e
	}

	fmt.Printf("%s", string(body))
	todo := new(ICloudRemindersApp)
	if e = json.Unmarshal(body, todo); e != nil {
		fmt.Printf("%# v", pretty.Formatter(todo))
		return nil, e
	}

	todo.GUID2Name = make(map[string]string)
	todo.Name2GUID = make(map[string]string)
	for _, c := range todo.Collections {
		todo.Name2GUID[c.Title] = c.GUID
		todo.GUID2Name[c.GUID] = c.Title
	}

	return todo, nil

}

type _collection struct {
	Ctag string `json:"ctag"`
	Guid string `json:"guid"`
}

type _clientState struct {
	Reminders   *ICloudReminder `json:"Reminders"`
	ClientState struct {
		Collections []_collection `json:"Collections"`
	} `json:"ClientState"`
}

func (app *ICloudRemindersApp) sync() {
	var a _clientState
	a.ClientState.Collections = make([]_collection, len(app.Collections))
	for j, i := range app.Collections {
		a.ClientState.Collections[j] = _collection{i.Ctag, i.GUID}
	}
	a.Reminders, _ = app.NewReminder("test", "shopping")

	b, _ := json.Marshal(a)
	fmt.Print(string(b))
}

// Return a new empty reminder. Many of these fields will be filled in on return from icloud
func (app *ICloudRemindersApp) NewReminder(title, parent string) (*ICloudReminder, error) {
	var ok bool
	rval := new(ICloudReminder)
	rval.CreatedDateExtended = time.Now().Unix()

	// 970378189: magic 20001001, some Apple history date?
	if rval.PGUID, ok = app.Name2GUID[parent]; !ok {
		rval.PGUID = "tasks"
	}
	rval.GUID = strings.ToUpper(uuid.NewRandom().String())
	rval.Title = title
	return rval, nil
}
