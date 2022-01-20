package openplatform

import (
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/wechat/util"
)

const urlStartPushTicket = "https://api.weixin.qq.com/cgi-bin/component/api_start_push_ticket"

// StartPushTicket 启动票据推送服务。
// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/ThirdParty/token/component_verify_ticket_service.html
func (open *OpenPlatform) StartPushTicket() *util.WechatError {
	_, err := util.PostJSON(urlStartPushTicket, g.Map{
		"component_appid":  open.Context.AppID,
		"component_secret": open.Context.AppSecret,
	})
	if err != nil {
		return util.UnknownError(err)
	}
	return nil
}
