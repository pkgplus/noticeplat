package wechat

import (
	"fmt"
	"log"
	"os"

	"errors"
	"github.com/esap/wechat"
	"github.com/esap/wechat/util"
)

const (
	FMT_URL_JSCODE2SESSION = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
)

var (
	WxServer *wechat.Server
)

func init() {
	appid := os.Getenv("WX_APPID")
	secret := os.Getenv("WX_SECRET")
	if appid == "" || secret == "" {
		panic("can't get WX_APPID and WX_SECRET from env")
	}

	WxServer = wechat.New("", appid, secret, "")
	WxServer.MsgUrl = "https://api.weixin.qq.com/cgi-bin/message/wxopen/template/send?access_token="
}

func SendMsg(t *TemplateMsg) error {
	// log.Println("prepare send message...")
	wxErr := WxServer.SendMsg(t)
	if wxErr.Error() != nil {
		log.Println(wxErr.ErrMsg)
	}

	return wxErr.Error()
}

type SessionResp struct {
	OpenID  string `json:"openid"`
	SessKey string `json:"session_key"`
	Unionid string `json:"unionid"`
}

func NewSessionResp(sessMap map[string]string) (*SessionResp, error) {
	openid := sessMap["openid"]
	session_key := sessMap["session_key"]
	unionid := sessMap["unionid"]

	if openid == "" || session_key == "" {
		return nil, errors.New("openid/session_key must not null")
	}

	return &SessionResp{openid, session_key, unionid}, nil
}

func (s *SessionResp) Convert2Map() map[string]interface{} {
	return map[string]interface{}{
		"openid":      s.OpenID,
		"session_key": s.SessKey,
		"unionid":     s.Unionid,
	}
}

func Jscode2Session(jscode string) (sessResp *SessionResp, err error) {
	req_url := fmt.Sprintf(FMT_URL_JSCODE2SESSION,
		WxServer.AppId,
		WxServer.Secret,
		jscode)

	sessResp = new(SessionResp)
	err = util.GetJson(req_url, sessResp)
	return
}
