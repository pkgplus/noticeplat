package hrsign

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bingbaba/hhsecret/web"
	"github.com/xuebing1110/noticeplat/user"
)

const (
	PLUGIN_TYPE_HRSIGN         = "HrSign"
	FMT_HRSIGN_NOTICECHECK_URL = "https://m.bingbaba.com/api/user/%s/notice"
)

var (
	ERR_NOTICE_CHECK = errors.New("CheckNoticeFailed")
)

type HrSignPlugin struct {
	HrUserID string
}

func (hs *HrSignPlugin) GetType() string {
	return PLUGIN_TYPE_HRSIGN
}
func (hs *HrSignPlugin) Execute(ups *user.UserPluginSetting) error {
	real_url := fmt.Sprintf(FMT_HRSIGN_NOTICECHECK_URL, hs.HrUserID)
	resp, err := http.DefaultClient.Get(real_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	notice_resp := new(web.Response)
	err = json.Unmarshal(data, notice_resp)
	if err != nil {
		return err
	}

	flag, ok := notice_resp.Data.(bool)
	if !ok {
		return ERR_NOTICE_CHECK
	}

	if flag {
		log.Printf("should send a notice: %s!", ups.String())
	} else {
		log.Printf("can't send notice: %s!", ups.String())
	}
	return nil
}
func (hs *HrSignPlugin) GetTemplateMsgID() string {
	return "8U98v1g7PWLZ5p4jbWNSpY5dr-hhG5kVuMAUew4PHnY"
}
