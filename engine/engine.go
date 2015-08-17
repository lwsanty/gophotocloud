package engine

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	//	"strings"
	//	"github.com/dyy18/istreamdatago/util"
)

// Types global to ICloud

type ICloudService struct {
	Name        string
	Url         string `json:"url"`
	Status      string `json:"status"`
	PcsRequired bool   `json:"pcsRequired"`
}

type ICloudUser struct {
	AppleIdAliases            []string `json:"appleIdAliases"`
	FirstName                 string   `json:"firtName"`
	FullName                  string   `json:"fullName"`
	Locked                    bool     `json:"locked"`
	ADsID                     string   `json:"aDsID"` // This is a UUID
	LanguageCode              string   `json:"languageCode"`
	BrMigrated                bool     `json:"brMigrated"`
	StatusCode                uint32   `json:"statusCode"`
	PrimaryEmail              string   `json:"primaryEmail"`
	Dsid                      string   `json:"dsid"`
	GilliganEnabled           bool     `json:"gilligan-enabled"`
	GilliganInvited           bool     `json:"gilligan-invited"`
	AppleId                   string   `json:"appleId"`
	IsPaidDeveloper           bool     `json:"isPaidDeveloper"`
	HasICloudQualifyingDevice bool     `json:"hasICloudQualifyingDevice"`
	Locale                    string   `json:"locale"`
	AppleIdAlias              string   `json:"appleIdAlias"`
	LastName                  string   `json:"lastName"`
	ICloudAppleIdAlias        string   `json:"iCloudAppleIdAlias"`
	PrimaryEmailVerified      bool     `json:"primaryEmailVerified"`
}

type ICloudEngine struct {
	Client          *http.Client
	ClientID        string
	ReportedVersion *ICloudVersion
	Cookiez         []*http.Cookie
	Version         uint32 `json:"version"`
	PcsEnabled      bool   `json:"pcsEnabled"`
	RequestInfo     struct {
		Country  string `json:"country"`
		TimeZone string `json:"timeZone"`
		Region   string `json:"region"`
	} `json:"requestInfo"`
	HasMinimumDeviceForPhotosWeb bool                     `json:"hasMinimumDeviceForPhotosWeb"`
	Apps                         map[string]string        `json:"apps"`
	PcsServiceIdentitiesIncluded bool                     `json:"pcsServiceIdentitiesIncluded"`
	AppsOrder                    []string                 `json:"appsOrder"`
	User                         ICloudUser               `json:"dsInfo"`
	Webservices                  map[string]ICloudService `json:"webservices"`
	IsExtendedLogin              bool                     `json:"isExtendedLogin"`
}

type ICloudVersion struct {
	AutoUpdate  string `json:"autoUpdate"`
	BuildNumber string `json:"buildNumber"`
}

/* ClientID is a UUID that seems to be arbitrary. This is the ClientID for this Library */
const ClientID = "3F93B25B-E569-4A3B-A1BC-022D4C19BF4C"

const versionURL = "https://www.icloud.com/system/cloudos/current/version.json"
const loginURL = "https://setup.icloud.com/setup/ws/1/login"

func getICloudVersion(client *http.Client) (*ICloudVersion, error) {
	var req *http.Request
	var e error

	if req, e = http.NewRequest("GET", versionURL, nil); e != nil {
		return nil, e
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Add("Origin", "https://www.icloud.com")

	var resp *http.Response
	if resp, e = client.Do(req); e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(resp.Body); e != nil {
		return nil, e
	}
	version := &ICloudVersion{}
	if e = json.Unmarshal(body, version); e != nil {
		return nil, e
	}
	return version, nil
}

// Functions from here on are exported

// Functions exported on the ICloudEngine type....

func NewEngine(apple_id, password string) (engine *ICloudEngine, e error) {
	engine = new(ICloudEngine)

	engine.ClientID = ClientID //strings.ToUpper(util.GenUuid())

	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: cookieJar,
	}

	var version *ICloudVersion
	if version, e = getICloudVersion(client); e != nil {
		return nil, e
	}

	engine.ReportedVersion = version
	info := map[string]string{
		"apple_id":       apple_id,
		"password":       password,
		"extended_login": "false",
	}

	data, _ := json.Marshal(info)

	var req *http.Request
	if req, e = http.NewRequest("POST", loginURL, bytes.NewReader(data)); e != nil {
		return nil, e
	}

	v := url.Values{}
	v.Add("clientBuildNumber", version.BuildNumber)
	v.Add("clientID", engine.ClientID)
	req.URL.RawQuery = v.Encode()
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Add("Origin", "https://www.icloud.com")

	var resp *http.Response
	if resp, e = client.Do(req); e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(resp.Body); e != nil {
		return nil, e
	}

	engine.Cookiez = resp.Cookies()

	json.Unmarshal(body, engine)
	engine.Client = client
	return engine, nil
}
