package photos

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"encoding/json"
	"github.com/icloud/download"
	"github.com/mig2/icloud/engine"
	"strings"
	"strconv"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type PhotosResp struct {
	SyncToken         string `json:"syncToken"`
	AssetsBinaryFeed  string `json:"assetsBinaryFeed"`
	MomentsBinaryFeed string `json:"momentsBinaryFeed"`
	IsUploadEnabled   bool   `json:"isUploadEnabled"`
}

type Assets struct {
	Assets []Asset `json:"assets"`
}

type Asset struct {
	ClientId       int      `json:"clientId"`
	CreatedDate    int      `json:"createdDate"`
	SubTitleTokens []string `json:"subTitleTokens"`
	Details        Detail   `json:"details"`
	Etag           string   `json:"etag"`
	IsEdited       bool     `json:"isEdited"`
	Type           string   `json:"type"`
	ServerId       string   `json:"serverId"`
	TitleTokens    []string `json:"titleTokens"`
	IsFavorite     bool     `json:"isFavorite"`
	Dimensions     []int    `json:"dimensions"`
	DerivativeInfo []string `json:"derivativeInfo"`
}

type KeyAssetIds struct {
	Kids []KeyAssetId `json:"folders"`
}

type KeyAssetId struct {
	KeyAssetClientId int `json:"keyAssetClientId"`
}

type Detail struct {
	Filename string `json:"filename"`
	Filesize int    `json:"filesize"`
}

func MaxValue(arr []int) (int, error) {
	if len(arr) == 0 {
		return 0, Error("empty")
	}

	max := arr[0]

	for _, value := range arr {
		if value > max {
			max = value
		}
	}

	return max, nil
}

func ClientIds(last int) string {
	//[1, 2, 3, 4, 5]"
	s := "["
	for i := 1; i <= last; i++ {
		s += strconv.Itoa(i)
		if i != last {
			s += ", "
		}
	}
	s += "]"

	return s
}

func NewP(cloud *engine.ICloudEngine) error {
	var svc engine.ICloudService
	var e error
	var ok bool
	if cloud.Client == nil {
		return Error("Client not logged in")
	}
	if svc, ok = cloud.Webservices["photos"]; !ok {
		return Error("No Photos")
	}
	if svc.Status != "active" {
		return Error("Photos service not active")
	}

	host, _, _ := net.SplitHostPort(svc.Url)
	var req *http.Request
	if req, e = http.NewRequest("GET", host+"/ph/startup", nil); e != nil {
		return e
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
		return e
	}

	defer resp.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(resp.Body); e != nil {
		return e
	}

	fmt.Printf("%s", string(body))

	//===========================================================================================================

	presp := new(PhotosResp)
	if e = json.Unmarshal(body, &presp); e != nil {
		return e
	}

	fmt.Printf("\n===========================================\n")

	var req1 *http.Request
	if req1, e = http.NewRequest("GET", host+"/ph/folders", nil); e != nil {
		return e
	}

	v.Add("syncToken", presp.SyncToken)

	req1.URL.RawQuery = v.Encode()
	req1.Header.Set("Content-Type", "text/plain")
	req1.Header.Add("Origin", "https://www.icloud.com")

	var resp1 *http.Response
	if resp1, e = cloud.Client.Do(req1); e != nil {
		return e
	}

	defer resp1.Body.Close()

	var body1 []byte
	if body1, e = ioutil.ReadAll(resp1.Body); e != nil {
		return e
	}

	keyid := new(KeyAssetIds)
	if e = json.Unmarshal(body1, &keyid); e != nil {
		return e
	}

	keyids := make([]int, len(keyid.Kids))

	for i := range keyid.Kids {
		keyids[i] = int(keyid.Kids[i].KeyAssetClientId)
	}

	maxId, err := MaxValue(keyids)
	if err != nil {
		return err
	}

	clientIds := ClientIds(maxId)
	fmt.Printf("\n===========================================\n")
	var req_assets *http.Request
	if req_assets, e = http.NewRequest("GET", host+"/ph/assets", nil); e != nil {
		return e
	}

	v.Add("methodOverride", "GET")
	v.Add("clientIds", clientIds)

	req_assets.URL.RawQuery = v.Encode()
	req_assets.Header.Set("Content-Type", "text/plain")
	req_assets.Header.Add("Origin", "https://www.icloud.com")

	var resp_assets *http.Response
	if resp_assets, e = cloud.Client.Do(req_assets); e != nil {
		return e
	}

	defer resp_assets.Body.Close()

	var body_assets []byte
	if body_assets, e = ioutil.ReadAll(resp_assets.Body); e != nil {
		return e
	}

	fmt.Printf("%s", string(body_assets))

	asst := new(Assets)
	if e = json.Unmarshal(body_assets, &asst); e != nil {
		return e
	}

	fmt.Printf("\n===========================================\n")

	direct_links := make([]string, len(asst.Assets))

	for i := range asst.Assets {
		s := asst.Assets[i].DerivativeInfo[0]
		ind := strings.Index(s, "https")
		decoded, err := url.QueryUnescape(s[ind:])
		if err != nil {
			return err
		}
		direct_links[i] = decoded
		fmt.Println(direct_links[i])

		filename := asst.Assets[i].Details.Filename
		fmt.Printf("\n===========================================\n")
		download.DownloadFromUrl(direct_links[i], filename)
	}

	return nil
}


