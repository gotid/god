package openplatform

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"time"

	"git.zc0901.com/go/god/lib/wechat/openplatform/context"

	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/wechat/msg"
	"git.zc0901.com/go/god/lib/wechat/util"
)

var (
	xmlContentType   = []string{"application/xml; charset=utf-8"}
	plainContentType = []string{"text/plain; charset=utf-8"}
)

// Server 开放平台 http 服务器。
type Server struct {
	*context.Context
	Request *http.Request
	Writer  http.ResponseWriter

	messageHandler func(*msg.Msg) (*msg.Response, error) // 消息回调

	RequestRaw  []byte      // 微信请求原始XML消息
	ResponseRaw []byte      // 返给微信的原始XML信息
	RequestMsg  *msg.Msg    // 解析后的微信请求消息
	ResponseMsg interface{} // 返给微信的响应信息

	skipValidate bool   // 是否跳过签名验证
	isSafeMode   bool   // 是否为安全模式
	openID       string // 请求 openID
	timestamp    int64
	nonce        string
	random       []byte
}

// GetServer 消息管理：接收第三方平台事件，被动回复消息管理
func GetServer(ctx *context.Context, w http.ResponseWriter, r *http.Request) *Server {
	s := &Server{
		Context: ctx,
		Writer:  w,
		Request: r,
	}
	return s
}

// SetMessageHandler 设置消息响应处理器
func (s *Server) SetMessageHandler(handler func(*msg.Msg) (*msg.Response, error)) {
	s.messageHandler = handler
}

// Serve 处理微信的请求信息。
func (s *Server) Serve() error {
	// 校验签名
	if !s.Validate() {
		return fmt.Errorf("请求校验失败")
	}

	// 响应 echostr
	echoStr, exists := s.GetQuery("echostr")
	if exists {
		s.String(echoStr)
		return nil
	}

	// 处理请求
	resp, err := s.handleRequest()
	if err != nil {
		return err
	}

	// 构建响应
	return s.buildResponse(resp)
}

// Send 返回自定义消息给微信
func (s *Server) Send() (err error) {
	replyMsg := s.ResponseMsg
	logx.Debugf("响应信息 %+v", replyMsg)

	// 安全模式需加密
	if s.isSafeMode {
		var encryptedMsg []byte
		encryptedMsg, err = util.EncryptMsg(s.random, s.ResponseRaw, s.AppID, s.EncodingAESKey)
		if err != nil {
			return
		}
		// TODO 如果获取不到timestamp nonce 则自己生成
		timestamp := s.timestamp
		timestampStr := strconv.FormatInt(timestamp, 10)
		msgSignature := util.Signature(s.Token, timestampStr, s.nonce, string(encryptedMsg))
		replyMsg = msg.EncryptedResponseMsg{
			EncryptedMsg: string(encryptedMsg),
			MsgSignature: msgSignature,
			Timestamp:    timestamp,
			Nonce:        s.nonce,
		}
	}

	// 返给微信 XML 响应
	if replyMsg != nil {
		s.XML(replyMsg)
	}

	return nil
}

// Validate 校验请求是否合法
func (s *Server) Validate() bool {
	if s.skipValidate {
		return true
	}

	timestamp := s.Query("timestamp")
	nonce := s.Query("nonce")
	signature := s.Query("signature")
	logx.Debugf("验证签名：timestamp=%s, nonce=%s", timestamp, nonce)
	return signature == util.Signature(s.Token, timestamp, nonce)
}

// Query 返回网址中查询键的值。
func (s *Server) Query(key string) string {
	v, _ := s.GetQuery(key)
	return v
}

// GetQuery 返回网址中查询键的值及存在状态。
func (s *Server) GetQuery(key string) (string, bool) {
	if vs, ok := s.Request.URL.Query()[key]; ok && len(vs) > 0 {
		return vs[0], true
	}

	return "", false
}

// 输出字符串
func (s *Server) String(str string) {
	s.SetContentType(s.Writer, plainContentType)
	s.WriteBytes([]byte(str))
}

// XML 输出 XML
func (s *Server) XML(obj interface{}) {
	s.SetContentType(s.Writer, xmlContentType)
	bytes, err := xml.Marshal(obj)
	if err != nil {
		panic(err)
	}

	s.WriteBytes(bytes)
}

// WriteBytes 输出字节流
func (s *Server) WriteBytes(bytes []byte) {
	s.Writer.WriteHeader(200)
	_, err := s.Writer.Write(bytes)
	if err != nil {
		panic(err)
	}
}

// SetContentType 设置内容类型。
func (s *Server) SetContentType(w http.ResponseWriter, contentType []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = contentType
	}
}

// 处理微信请求信息
func (s *Server) handleRequest() (reply *msg.Response, err error) {
	// 设置安全模式
	if v := s.Query("encrypt_type"); v == "aes" {
		s.isSafeMode = true
	} else {
		s.isSafeMode = false
	}

	// 设置 openid
	s.openID = s.Query("openid")

	// 获取微信请求信息
	m, err := s.getMessage()
	if err != nil {
		return nil, err
	}
	s.RequestMsg = m

	return s.messageHandler(m)
}

// 构建返给微信的响应
func (s *Server) buildResponse(resp *msg.Response) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic error: %v\n%s", e, debug.Stack())
		}
	}()

	// 跳过空回复
	if resp == nil {
		return nil
	}

	// 处理不支持的消息类型
	switch resp.Scene {
	case msg.ResponseSceneOpen:
		s.buildOpenReply(resp)
	case msg.ResponseSceneKefu:
		// s.buildKefuReply(resp)
	case msg.ResponseScenePay:
		// s.buildPayReply(resp)
	}

	switch resp.Type {
	case msg.TypeText:
	case msg.TypeImage:
	case msg.TypeVoice:
	case msg.TypeVideo:
	case msg.TypeMusic:
	case msg.TypeNews:
	case msg.TypeTransferKf:
	default:
		err = msg.ErrUnsupportedResponse
		return
	}

	// 构建回复数据
	data := resp.Msg
	value := reflect.ValueOf(data)
	kind := value.Kind().String()

	// 处理非指针类型的数据值
	if kind != "ptr" {
		return msg.ErrUnsupportedResponse
	}

	// 赋值给通用基础信息
	params := make([]reflect.Value, 1)
	params[0] = reflect.ValueOf(s.RequestMsg.FromUserName)
	value.MethodByName("SetToUserName").Call(params)

	params[0] = reflect.ValueOf(s.RequestMsg.ToUserName)
	value.MethodByName("SetFromUserName").Call(params)

	params[0] = reflect.ValueOf(resp.Type)
	value.MethodByName("SetMsgType").Call(params)

	params[0] = reflect.ValueOf(time.Now().Unix())
	value.MethodByName("SetCreateTime").Call(params)

	s.ResponseMsg = data
	s.ResponseRaw, err = xml.Marshal(data)
	return
}

// 获取解析后的微信信息
func (s *Server) getMessage() (*msg.Msg, error) {
	var raw []byte
	var err error

	if s.isSafeMode {
		var encryptedMsg msg.EncryptedMsg
		if err := xml.NewDecoder(s.Request.Body).Decode(&encryptedMsg); err != nil {
			return nil, fmt.Errorf("解析微信请求消息体失败：err=%v", err)
		}

		// 验证加密消息签名
		timestamp := s.Query("timestamp")
		s.timestamp, _ = strconv.ParseInt(timestamp, 10, 32)
		if err != nil {
			return nil, err
		}
		s.nonce = s.Query("nonce")
		msgSignature := s.Query("msg_signature")
		if msgSignature != util.Signature(s.Token, timestamp, s.nonce, encryptedMsg.Encrypt) {
			return nil, fmt.Errorf("加密请求校验失败")
		}

		// 解密消息
		s.random, raw, err = util.DecryptMsg(s.AppID, encryptedMsg.Encrypt, s.EncodingAESKey)
		if err != nil {
			return nil, fmt.Errorf("加密消息体解密失败, err=%v", err)
		}
	} else {
		raw, err = ioutil.ReadAll(s.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("解析微信请求消息体失败：err=%v", err)
		}
	}

	// 原始请求消息体
	s.RequestRaw = raw

	return s.parseRequestMessage(raw)
}

// 解析微信请求信息
func (s *Server) parseRequestMessage(raw []byte) (m *msg.Msg, err error) {
	m = &msg.Msg{}
	if err = xml.Unmarshal(raw, m); err != nil {
		return nil, err
	}

	return m, nil
}

// 构建开放平台回复消息
func (s *Server) buildOpenReply(reply *msg.Response) {
	// 设置默认响应值
	if reply.Type == "" {
		reply.Type = msg.ResponseTypeString
	}
	if reply.Msg == nil {
		reply.Msg = "success"
	}

	if s.RequestMsg.InfoType == msg.InfoTypeVerifyTicket {
		// s.Context
	}
}
