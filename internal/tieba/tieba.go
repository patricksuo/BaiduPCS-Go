// Package tieba 通过百度贴吧客户端接口校验 BDUSS, 并获取百度账号的 UID / 用户名 / 性别 / 吧龄等信息。
//
// 这部分逻辑原本依赖外部模块 github.com/qjfoidnh/baidu-tools (tieba / tiebautil / randominfo 三个子包),
// 现已 port 进本项目以去除该依赖 (顺带可移除 github.com/bitly/go-simplejson)。行为与原实现保持一致。
package tieba

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/qjfoidnh/BaiduPCS-Go/requester"
	"github.com/tidwall/gjson"
)

// 贴吧客户端签名固定后缀
const tiebaSignSuffix = "tiebaclient!!!"

// Baidu 百度账号详细情况
type Baidu struct {
	UID      uint64  // 百度ID对应的uid
	Name     string  // 真实ID(用户名)
	NameShow string  // 显示的用户名(昵称)
	Sex      string  // 性别
	Age      float64 // 账号年龄(贴吧吧龄)
}

// Tieba 百度贴吧账号详细情况
type Tieba struct {
	Baidu *Baidu
	Tbs   string
}

// sumIMEI 根据 key 计算出 imei (port 自 baidu-tools/randominfo)
func sumIMEI(key string) uint64 {
	var hash uint64 = 53202347234687234
	for k := range key {
		hash += (hash << 5) + uint64(key[k])
	}
	hash %= uint64(1e15)
	if hash < 1e14 {
		hash += 1e14
	}
	return hash
}

// getPhoneModel 根据 key, 从 phoneModelDataBase 中取出一个稳定的手机型号 (port 自 baidu-tools/randominfo)
func getPhoneModel(key string) string {
	if len(phoneModelDataBase) <= 0 {
		return "S3"
	}
	var hash uint64 = 2134
	for k := range key {
		hash += (hash << 4) + uint64(key[k])
	}
	hash %= uint64(len(phoneModelDataBase))
	return phoneModelDataBase[int(hash)]
}

// stringReverse 翻转字符串
func stringReverse(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// tiebaClientSignature 按贴吧客户端规则给 post 数据补全设备字段并签名 (port 自 baidu-tools/tiebautil).
// 注意: 调用方一般不会在 post 中放入 "BDUSS", 此时设备信息退化为一个固定值, 与原实现一致。
func tiebaClientSignature(post map[string]string) {
	if post == nil {
		return
	}
	delete(post, "sign")

	var (
		bduss        = post["BDUSS"]
		model        = getPhoneModel(bduss)
		phoneIMEIStr = strconv.FormatUint(sumIMEI(model+"_"+bduss), 10)
		m            = md5.New()
	)

	post["_client_type"] = "2"
	post["_client_version"] = "7.0.0.0"
	post["_phone_imei"] = phoneIMEIStr
	post["from"] = "mini_ad_wandoujia"
	post["model"] = model
	m.Write([]byte(bduss + "_" + post["_client_version"] + "_" + post["_phone_imei"] + "_" + post["from"]))
	post["cuid"] = strings.ToUpper(hex.EncodeToString(m.Sum(nil))) + "|" + stringReverse(phoneIMEIStr)

	keys := make([]string, 0, len(post))
	for key := range post {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	m.Reset()
	for _, key := range keys {
		m.Write([]byte(key + "=" + post[key]))
	}
	m.Write([]byte(tiebaSignSuffix))
	post["sign"] = strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}

// tiebaClientRawQuerySignature 给 rawQuery 进行贴吧客户端签名 (port 自 baidu-tools/tiebautil)
func tiebaClientRawQuerySignature(rawQuery string) string {
	m := md5.New()
	m.Write([]byte(strings.ReplaceAll(rawQuery, "&", "")))
	m.Write([]byte(tiebaSignSuffix))
	return rawQuery + "&sign=" + strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}

func parseBody(r io.Reader, what string) (gjson.Result, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("%s读取响应出错: %s", what, err)
	}
	if !gjson.ValidBytes(body) {
		return gjson.Result{}, fmt.Errorf("%s json解析出错", what)
	}
	return gjson.ParseBytes(body), nil
}

func sexString(sex int64) string {
	switch sex {
	case 1:
		return "♂"
	case 2:
		return "♀"
	default:
		return "unknown"
	}
}

// NewUserInfoByBDUSS 检测 BDUSS 有效性, 同时获取百度账号详细信息 (无法获取 ptoken 和 stoken)。
func NewUserInfoByBDUSS(bduss string) (*Tieba, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	post := map[string]string{
		"bdusstoken":  bduss + "|null",
		"channel_id":  "",
		"channel_uid": "",
		"stErrorNums": "0",
		"subapp_type": "mini",
		"timestamp":   timestamp + "922",
	}
	tiebaClientSignature(post)

	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Cookie":       "ka=open",
		"net":          "1",
		"User-Agent":   "bdtb for Android 6.9.2.1",
		"client_logid": timestamp + "416",
		"Connection":   "Keep-Alive",
	}

	// 获取百度ID的 TBS, UID, BDUSS 等
	resp, err := requester.DefaultClient.Req("POST", "http://tieba.baidu.com/c/s/login", post, header)
	if err != nil {
		return nil, fmt.Errorf("检测BDUSS有效性网络错误, %s", err)
	}
	defer resp.Body.Close()

	json, err := parseBody(resp.Body, "检测BDUSS有效性")
	if err != nil {
		return nil, err
	}

	if errCode := json.Get("error_code").String(); errCode != "0" {
		return nil, fmt.Errorf("检测BDUSS有效性错误代码: %s, 消息: %s", errCode, json.Get("error_msg").String())
	}

	uid, _ := strconv.ParseUint(json.Get("user.id").String(), 10, 64)
	t := &Tieba{
		Baidu: &Baidu{
			UID:  uid,
			Name: json.Get("user.name").String(),
		},
		Tbs: json.Get("anti.tbs").String(),
	}

	if err = t.FlushUserInfo(); err != nil {
		return nil, err
	}
	return t, nil
}

// NewUserInfoByUID 提供 UID 获取百度账号详细信息。
func NewUserInfoByUID(uid uint64) (*Tieba, error) {
	rawQuery := "has_plist=0&need_post_count=1&rn=1&uid=" + strconv.FormatUint(uid, 10)
	urlStr := "http://c.tieba.baidu.com/c/u/user/profile?" + tiebaClientRawQuerySignature(rawQuery)
	resp, err := requester.DefaultClient.Req("GET", urlStr, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	json, err := parseBody(resp.Body, "获取百度账号信息")
	if err != nil {
		return nil, err
	}

	user := json.Get("user")
	age, _ := strconv.ParseFloat(user.Get("tb_age").String(), 64)
	return &Tieba{
		Baidu: &Baidu{
			UID:      uid,
			Name:     user.Get("name").String(),
			NameShow: user.Get("name_show").String(),
			Sex:      sexString(user.Get("sex").Int()),
			Age:      age,
		},
	}, nil
}

// NewUserInfoByName 提供 name (百度用户名) 获取百度账号详细信息。
func NewUserInfoByName(name string) (*Tieba, error) {
	resp, err := requester.DefaultClient.Req("GET", "http://tieba.baidu.com/home/get/panel?un="+name, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	json, err := parseBody(resp.Body, "获取百度账号信息")
	if err != nil {
		return nil, err
	}
	return NewUserInfoByUID(json.Get("data.id").Uint())
}

// FlushUserInfo 刷新百度账号详细信息。
func (t *Tieba) FlushUserInfo() error {
	if t.Baidu == nil {
		return fmt.Errorf("Baidu is not initialize")
	}

	var (
		fresh *Tieba
		err   error
	)
	switch {
	case t.Baidu.UID != 0:
		fresh, err = NewUserInfoByUID(t.Baidu.UID)
	case t.Baidu.Name != "":
		fresh, err = NewUserInfoByName(t.Baidu.Name)
	default:
		return fmt.Errorf("Baidu uid and name are null")
	}
	if err != nil {
		return err
	}
	t.Baidu = fresh.Baidu
	return nil
}
