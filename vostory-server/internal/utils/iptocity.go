package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/axgle/mahonia"
)

// IsPrivateIP 判断是否为内网IP
func IsPrivateIP(ipStr string) bool {
	// 将字符串解析为 net.IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false // 如果解析失败，返回 false
	}

	// 定义内网IP地址段
	privateIPBlocks := []*net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
		{IP: net.ParseIP("127.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)},
		{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},
	}

	// 遍历所有内网地址段，检查IP是否在其中
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

func GetCityByIp(ip string) (string, error) {
	if ip == "" {
		return "", errors.New("ip为空")
	}
	if IsPrivateIP(ip) {
		return "内网IP", nil
	}

	//美团IP地址查询
	//https://apimobile.meituan.com/locate/v2/ip/loc?rgeo=true&ip=123.123.123.123

	url := "http://whois.pconline.com.cn/ipJson.jsp?json=true&ip=" + ip
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Accept-Charset", "GBK,utf-8;q=0.7,*;q=0.3")
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	body, _ := ioutil.ReadAll(response.Body)
	bodystr := string(body)
	// tmp := cast.ToString(bodystr)
	tmp := ConvertToString(bodystr, "gbk", "utf-8")
	if response.StatusCode == 200 {
		p := make(map[string]interface{}, 0)
		if err := json.Unmarshal([]byte(tmp), &p); err == nil {
			return p["pro"].(string) + " " + p["city"].(string), nil
		}
	}
	return "", errors.New(tmp)
}

// src 字符串
// srcCode 字符串当前编码
// tagCode 要转换的编码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}
