package drive

import (
	"fmt"
	//"io/ioutil"
	//"net"
	//"net/http"
	//"net/url"

	//"encoding/json"
	"github.com/lwsanty/gophotocloud/engine"
	"net"
	"net/http"
	//	"net/url"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
)

type Objs struct {
	ob []Obj
}

type Obj struct {
	Drivewsid string `json:"drivewsid"`
	Docwsid   string `json:"docwsid"`
}

type Contents struct {
	cont []Content
}

type Content struct {
	Drivewsid     string `json:"drivewsid"`
	Docwsid       string `json:"docwsid"`
	Zone          string `json:"zone"`
	Name          string `json:"name"`
	Etag          string `json:"etag"`
	Type          string `json:"type"`
	Items         []Item `json:"items"`
	NumberOfItems int    `json:"numberOfItems"`
}

type Item struct {
	Drivewsid string `json:"drivewsid"`
	Docwsid   string `json:"docwsid"`
	Etag      string `json:"etag"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type FolderItems struct {
	Items []FolderItem
}

type FolderItem struct {
	Id   string
	Name string
	Type string
	Url  string
}

type File struct {
	Document_id string   `json:"document_id"`
	Data_token  FileData `json:"data_token"`
	Double_etag string   `json:"double_etag"`
}

type FileData struct {
	Token               string `json:"Token"`
	Url                 string `json:"url"`
	Signature           string `json:"signature"`
	Wrapping_key        string `json:"wrapping_key"`
	Reference_signature string `json:"reference_signature"`
}

type ICloudDriveFiles struct {
	Urls []string
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func GetContentsFilesIds(presp []Content) []string {
	if len(presp) == 0 {
		return nil
	}
	var s []string
	for i := range presp {
		for j := range presp[i].Items {
			if strings.Contains(presp[i].Items[j].Drivewsid, "FILE") {
				s = append(s, presp[i].Items[j].Docwsid)
			}
		}
	}

	return s
}

func GetToken(cookie string) string {
	//cookie will be like X-APPLE-WEBAUTH-TOKEN=v=2:t=AQAAAABV0fS93iv_VlHkHyy37x-GeujIInXPSQM~; Path=/; Domain=icloud.com; HttpOnly; Secure
	tl := strings.Index(cookie, "t=")
	trimmed_left := cookie[tl:]
	tr := strings.Index(trimmed_left, "; P")
	trimmed_right := trimmed_left[0:tr]
	s := strings.Replace(trimmed_right, "t=", "", -1)
	return s
}

func GetFileItemsUrls(fitems *FolderItems, cloud *engine.ICloudEngine, cookie string, token string) (*FolderItems, error) {
	for i := range fitems.Items {
		fit, err := GetFileItemUrl(&fitems.Items[i], cloud, cookie , token);
		if err != nil {
			panic(err)
		}

		fitems.Items[i] = *fit
	}
	return fitems, nil
}

func GetFileItemUrl(fitem *FolderItem, cloud *engine.ICloudEngine, cookie string, token string) (*FolderItem, error) {
	var reqfile *http.Request
	var svdoc engine.ICloudService
	var e error
	var ok bool
	if svdoc, ok = cloud.Webservices["docws"]; !ok {
		return nil, Error("No drivews")
	}
	if svdoc.Status != "active" {
		return nil, Error("Dockws service is not active")
	}
	hostdoc, _, _ := net.SplitHostPort(svdoc.Url)
	if reqfile, e = http.NewRequest("GET", hostdoc+"/ws/com.apple.CloudDocs/download/by_id", nil); e != nil {
		return nil, e
	}

	v := url.Values{}
	v.Set("document_id", fitem.Id)
	v.Set("token", token)
	reqfile.URL.RawQuery = v.Encode()

	reqfile.Header.Set("Content-Type", "text/plain")
	reqfile.Header.Add("Cookie", cookie)
	reqfile.Header.Add("Origin", "https://www.icloud.com")

	var resp *http.Response
	if resp, e = cloud.Client.Do(reqfile); e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(resp.Body); e != nil {
		return nil, e
	}

	var file File
	if e = json.Unmarshal(body, &file); e != nil {
		return nil, e
	}

	fitem.Url = file.Data_token.Url

	fmt.Println("fitem.Url = ", fitem.Url)
	return fitem, nil
}

func GetFolderItems(cloud *engine.ICloudEngine, folder string) (*FolderItems, string, string, error) {
	var svc engine.ICloudService
	var svdoc engine.ICloudService
	var e error
	var ok bool
	if cloud.Client == nil {
		return nil, "", "", Error("Client not logged in")
	}

	//fmt.Print(cloud.Cookiez)

	//check photos from webservices
	if svc, ok = cloud.Webservices["drivews"]; !ok {
		return nil, "",  "", Error("No drivews")
	}
	if svc.Status != "active" {
		return nil, "", "", Error("Drivews service is not active")
	}

	if svdoc, ok = cloud.Webservices["docws"]; !ok {
		return nil, "", "", Error("No drivews")
	}
	if svdoc.Status != "active" {
		return nil, "", "", Error("Dockws service is not active")
	}

	jsonStr := []byte(`[{drivewsid: "FOLDER::com.apple.CloudDocs::` + folder + `", partialData: false}]`)

	fmt.Println(string(jsonStr))

	host, _, _ := net.SplitHostPort(svc.Url)
	var req *http.Request
	/*
		for i:= range cloud.Cookiez {
			req.AddCookie(cloud.Cookiez[i])
		} */
	if req, e = http.NewRequest("POST", host+"/retrieveItemDetailsInFolders", bytes.NewBuffer(jsonStr)); e != nil {
		return nil, "", "", e
	}

	s := ""
	for i := range cloud.Cookiez {
		s += cloud.Cookiez[i].String()
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Add("Cookie", s)
	req.Header.Add("Origin", "https://www.icloud.com")

	var resp *http.Response
	if resp, e = cloud.Client.Do(req); e != nil {
		return nil, "", "", e
	}

	defer resp.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(resp.Body); e != nil {
		return nil, "", "", e
	}

	fmt.Println(string(body))

	fmt.Println("================================================================")

	//cookie will be like X-APPLE-WEBAUTH-TOKEN=v=2:t=AQAAAABV0fS93iv_VlHkHyy37x-GeujIInXPSQM~; Path=/; Domain=icloud.com; HttpOnly; Secure

	s += resp.Cookies()[0].String()

	token := GetToken(resp.Cookies()[0].String())

	fmt.Println("token: ", token)

	fmt.Println("================================================================")

	var presp []Content
	if e = json.Unmarshal(body, &presp); e != nil {
		return nil, "", "", e
	}

	if len(presp) == 0 {
		return nil, "", "", e
	}

	fitems := new(FolderItems)

	for i := range presp {
		for j := range presp[i].Items {
			fitems.Items = append(fitems.Items, FolderItem{Id:presp[i].Items[j].Docwsid, Name:presp[i].Items[j].Name, Type:presp[i].Items[j].Type})
		}
	}

	return fitems, s, token, nil

}

func NewD(cloud *engine.ICloudEngine) (*ICloudDriveFiles, error) {
	var svc engine.ICloudService
	var svdoc engine.ICloudService
	var e error
	var ok bool
	if cloud.Client == nil {
		return nil, Error("Client not logged in")
	}

	//fmt.Print(cloud.Cookiez)

	//check photos from webservices
	if svc, ok = cloud.Webservices["drivews"]; !ok {
		return nil, Error("No drivews")
	}
	if svc.Status != "active" {
		return nil, Error("Drivews service is not active")
	}

	if svdoc, ok = cloud.Webservices["docws"]; !ok {
		return nil, Error("No drivews")
	}
	if svdoc.Status != "active" {
		return nil, Error("Dockws service is not active")
	}

	jsonStr := []byte(`[{drivewsid: "FOLDER::com.apple.CloudDocs::root", partialData: false}]`)

	host, _, _ := net.SplitHostPort(svc.Url)
	hostdoc, _, _ := net.SplitHostPort(svdoc.Url)
	var req *http.Request
	/*
		for i:= range cloud.Cookiez {
			req.AddCookie(cloud.Cookiez[i])
		} */
	if req, e = http.NewRequest("POST", host+"/retrieveItemDetailsInFolders", bytes.NewBuffer(jsonStr)); e != nil {
		return nil, e
	}

	s := ""
	for i := range cloud.Cookiez {
		s += cloud.Cookiez[i].String()
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Add("Cookie", s)
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

	fmt.Println(string(body))

	fmt.Println("================================================================")

	//cookie will be like X-APPLE-WEBAUTH-TOKEN=v=2:t=AQAAAABV0fS93iv_VlHkHyy37x-GeujIInXPSQM~; Path=/; Domain=icloud.com; HttpOnly; Secure

	s += resp.Cookies()[0].String()

	token := GetToken(resp.Cookies()[0].String())

	fmt.Println("token: ", token)

	fmt.Println("================================================================")

	var presp []Content
	if e = json.Unmarshal(body, &presp); e != nil {
		return nil, e
	}

	fileIds := GetContentsFilesIds(presp)
	fmt.Println("id's", fileIds)

	iclouddrivefiles := new(ICloudDriveFiles)

	for i := range fileIds {
		var reqfile *http.Request

		if reqfile, e = http.NewRequest("GET", hostdoc+"/ws/com.apple.CloudDocs/download/by_id", nil); e != nil {
			return nil, e
		}

		v := url.Values{}
		v.Set("document_id", fileIds[i])
		v.Set("token", token)
		reqfile.URL.RawQuery = v.Encode()

		reqfile.Header.Set("Content-Type", "text/plain")
		reqfile.Header.Add("Cookie", s)
		reqfile.Header.Add("Origin", "https://www.icloud.com")

		var resp *http.Response
		if resp, e = cloud.Client.Do(reqfile); e != nil {
			return nil, e
		}

		defer resp.Body.Close()

		var body []byte
		if body, e = ioutil.ReadAll(resp.Body); e != nil {
			return nil, e
		}

		var file File
		if e = json.Unmarshal(body, &file); e != nil {
			return nil, e
		}

		iclouddrivefiles.Urls = append(iclouddrivefiles.Urls, file.Data_token.Url)
	}

	return iclouddrivefiles, nil
}
