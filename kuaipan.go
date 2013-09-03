package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type ErrorMsg struct {
	Msg string `json:"msg"`
}

func (e *ErrorMsg) Error() string {
	return e.Msg
}

type Kuaipan struct {
	root   string
	debug  bool
	client *Consumer
	atoken *AccessToken
}

func NewKuaipan(consumer_key, consumer_secret string) *Kuaipan {
	client := NewConsumer(
		consumer_key,
		consumer_secret,
		ServiceProvider{
			RequestTokenUrl:   "https://openapi.kuaipan.cn/open/requestToken",
			AuthorizeTokenUrl: "https://www.kuaipan.cn/api.php",
			AccessTokenUrl:    "https://openapi.kuaipan.cn/open/accessToken",
		})
	client.AdditionalAuthorizationUrlParams = map[string]string{
		"ac": "open",
		"op": "authorise",
	}
	return &Kuaipan{
		root:   "app_folder",
		client: client,
	}
}

func (k *Kuaipan) Debug(enabled bool) {
	k.debug = enabled
	k.client.Debug(enabled)
}

func (k *Kuaipan) Authorized() bool {
	return len(k.atoken.Token) > 0 && len(k.atoken.Secret) > 0
}

func (k *Kuaipan) SetAccessToken(token, secret string) {
	k.atoken = &AccessToken{Token: token, Secret: secret}
}

func (k *Kuaipan) GetAccessToken() (token, secret string) {
	return k.atoken.Token, k.atoken.Secret
}

func (k *Kuaipan) Authorize() bool {
	requestToken, url, err := k.client.GetRequestTokenAndUrl("")
	if err != nil {
		log.Fatal(err)
		return false
	}

	fmt.Println("(1) Go to: " + url)
	fmt.Println("(2) Grant access, you should get back a verification code.")
	fmt.Println("(3) Enter that verification code here: ")

	verificationCode := ""
	fmt.Scanln(&verificationCode)

	accessToken, err := k.client.AuthorizeToken(requestToken, verificationCode)
	if err != nil {
		log.Fatal(err)
		fmt.Print("授权失败")
		return false
	}
	fmt.Print("授权成功")
	k.atoken = accessToken
	return true
}

func getObject(resp *http.Response, obj interface{}) (err error) {
	if resp.StatusCode == 200 {
		if obj == nil {
			io.Copy(ioutil.Discard, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(obj)
		}
	} else {
		msg := new(ErrorMsg)
		if err = json.NewDecoder(resp.Body).Decode(msg); err == nil {
			err = msg
		} else {
			err = errors.New(resp.Status)
		}
	}

	return
}

type AccountInfo struct {
	UserId      int    `json:"user_id"`
	UserName    string `json:"user_name"`
	MaxFileSize int    `json:"max_file_size"`
	QuotaTotal  int64  `json:"quota_total"`
	QuotaUsed   int64  `json:"quota_used"`

	QuotoRecycled int64 `json:"quota_recycled"`
}

func (k *Kuaipan) AccountInfo() (*AccountInfo, error) {
	response, err := k.client.Get(
		"http://openapi.kuaipan.cn/1/account_info",
		map[string]string{},
		k.atoken)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	info := &AccountInfo{QuotoRecycled: -1}
	err = getObject(response, info)
	return info, err
}

type FileInfo struct {
	FileId     string `json:"file_id"`
	Type       string `json:"type"`
	Size       int    `json:"size"`
	CreateTime string `json:"create_time"`
	ModifyTime string `json:"modify_time"`
	Name       string `json:"name"`
	IsDeleted  bool   `json:"is_deleted"`

	Rev string `json:"rev"`
}

type DirInfo struct {
	Path string `json:"path"`
	Root string `json:"root"`

	Hash       string     `json:"hash"`
	FileId     string     `json:"file_id"`
	Type       string     `json:"type"`
	Size       int        `json:"size"`
	CreateTime string     `json:"create_time"`
	ModifyTime string     `json:"modify_time"`
	Name       string     `json:"name"`
	Rev        string     `json:"rev"`
	IsDeleted  bool       `json:"is_deleted"`
	Files      []FileInfo `json:"files"`
}

// 获取单个文件，文件夹信息
// params: list, file_limit, page, page_size, filter_ext, sort_by
func (k *Kuaipan) Metadata(pathname string, params map[string]string) (*DirInfo, error) {
	info := &DirInfo{Size: -1}
	response, err := k.client.Get(
		fmt.Sprintf("http://openapi.kuaipan.cn/1/metadata/%s/%s", k.root, pathname),
		params,
		k.atoken)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	err = getObject(response, info)
	return info, err
}

type ShareInfo struct {
	Url        string `json:"url"`
	AccessCode string `json:"access_code"`
}

// 创建并获取一个文件的分享链接
func (k *Kuaipan) Share(pathname, displayName, accessCode string) (*ShareInfo, error) {
	params := map[string]string{}
	if displayName != "" {
		params["name"] = displayName
	}
	if accessCode != "" {
		params["access_code"] = accessCode
	}
	//k.client.AdditionalAuthorizationUrlParams = params
	response, err := k.client.Get(
		fmt.Sprintf("http://openapi.kuaipan.cn/1/shares/%s/%s", k.root, pathname),
		params,
		k.atoken)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	res := new(ShareInfo)
	err = getObject(response, res)
	return res, err
}

type CreateResult struct {
	FileId string `json:"file_id"`
	Path   string `json:"path"`
	Root   string `json:"root"`
}

// 新建文件夹
func (k *Kuaipan) CreateFolder(pathname string) (*CreateResult, error) {
	res := new(CreateResult)
	response, err := k.client.Get(
		"http://openapi.kuaipan.cn/1/fileops/create_folder",
		map[string]string{"path": pathname, "root": k.root},
		k.atoken)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	getObject(response, res)
	return res, err
}

// 删除文件，文件夹，以及文件夹下所有文件到回收站
func (k *Kuaipan) Delete(pathname string, toRecycle bool) error {
	response, err := k.client.Get(
		"http://openapi.kuaipan.cn/1/fileops/delete",
		map[string]string{
			"path":       pathname,
			"root":       k.root,
			"to_recycle": strconv.FormatBool(toRecycle),
		},
		k.atoken)
	defer response.Body.Close()
	return err
}

// 移动文件，文件夹
func (k *Kuaipan) Move(fromPath, toPath string) error {
	response, err := k.client.Get(
		"http://openapi.kuaipan.cn/1/fileops/move",
		map[string]string{
			"from_path":       fromPath,
			"to_path":       toPath,
			"root":       k.root,
		},
		k.atoken)
	defer response.Body.Close()
	return err
}
