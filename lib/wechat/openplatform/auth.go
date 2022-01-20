package openplatform

import (
	"fmt"
	"net/url"
)

// 授权流程技术说明
// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Authorization_Process_Technical_Description.html

import (
	"encoding/json"

	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/wechat/util"
)

const (
	urlCreatePreAuthCode  = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token=%s"
	urlComponentLoginPage = "https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=%s&auth_type=%d&biz_appid=%s"
	urlBindComponent      = "https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&auth_type=%d&no_scan=1&component_appid=%s&pre_auth_code=%s&redirect_uri=%s&biz_appid=%s#wechat_redirect"
)

// PcAuthURL 获取PC版扫码授权链接
func (open *OpenPlatform) PcAuthURL(redirectURI string, authType int, bizAppID string) (string, error) {
	preAuthCode, err := open.PreAuthCode()
	if err != nil {
		return "", err
	}
	uri := url.QueryEscape(redirectURI)
	return fmt.Sprintf(urlComponentLoginPage, open.Context.AppID, preAuthCode, uri, authType, bizAppID), nil
}

// MobileAuthURL 获取移动版授权链接
func (open *OpenPlatform) MobileAuthURL(redirectURI string, authType int, bizAppID string) (string, error) {
	authCode, err := open.PreAuthCode()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(urlBindComponent, authType, open.Context.AppID, authCode, url.QueryEscape(redirectURI), bizAppID), nil
}

// PreAuthCode 获取预授权码
func (open *OpenPlatform) PreAuthCode() (string, error) {
	accessToken, err := open.AccessToken()
	if err != nil {
		return "", err
	}

	data, err := util.PostJSON(fmt.Sprintf(urlCreatePreAuthCode, accessToken), g.Map{
		"component_appid": open.Context.AppID,
	})
	if err != nil {
		return "", err
	}

	var ret struct {
		PreAuthCode string `json:"pre_auth_code"`
	}
	if err := json.Unmarshal(data, &ret); err != nil {
		return "", err
	}

	return ret.PreAuthCode, nil
}
