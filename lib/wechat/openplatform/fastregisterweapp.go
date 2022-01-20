package openplatform

import (
	"fmt"

	"git.zc0901.com/go/god/lib/wechat/util"
)

const (
	urlRegisterOrgWeApp = "https://api.weixin.qq.com/cgi-bin/component/fastregisterweapp?action=%s&component_access_token=%s"
)

// RegisterOrgWeAppParam 快速注册企业小程序参数
type RegisterOrgWeAppParam struct {
	Name               string `json:"name"`                 // 企业名
	Code               string `json:"code"`                 // 企业代码
	CodeType           string `json:"code_type"`            // 企业代码类型 1：统一社会信用代码（18 位） 2：组织机构代码（9 位 xxxxxxxx-x） 3：营业执照注册号(15 位)
	LegalPersonaWechat string `json:"legal_persona_wechat"` // 法人微信号
	LegalPersonaName   string `json:"legal_persona_name"`   // 法人姓名（绑定银行卡）
	ComponentPhone     string `json:"component_phone"`      // 第三方联系电话（方便法人与第三方联系）
}

// RegisterOrgWeApp 快速注册企业小程序
// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Register_Mini_Programs/Fast_Registration_Interface_document.html
func (open *OpenPlatform) RegisterOrgWeApp(param *RegisterOrgWeAppParam) error {
	accessToken, err := open.AccessToken()
	if err != nil {
		return err
	}

	url := fmt.Sprintf(urlRegisterOrgWeApp, "create", accessToken)
	data, err := util.PostJSON(url, param)
	if err != nil {
		return err
	}

	return util.TryDecodeError(data, "RegisterOrgWeApp")
}

// 查询企业小程序

// 注册个人小程序
// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Register_Mini_Programs/fastregisterpersonalweapp.html

// 查询个人小程序
