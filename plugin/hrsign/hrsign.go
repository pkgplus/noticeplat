package hrsign

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bingbaba/hhsecret"
	"github.com/bingbaba/hhsecret/web"
	"github.com/xuebing1110/noticeplat/plugin"
	"github.com/xuebing1110/noticeplat/user"
)

const (
	PLUGIN_TYPE_HRSIGN         = "HrSign"
	FMT_HRSIGN_NOTICECHECK_URL = "https://m.bingbaba.com/api/user/%s/notice"
	FMT_HRSIGN_SIGNLIST_URL    = "https://m.bingbaba.com/api/user/%s/sign"
)

var (
	ERR_NOTICE_CHECK = errors.New("CheckNoticeFailed")
)

func init() {
	err := plugin.Registe(PLUGIN_TYPE_HRSIGN, new(HrSignPlugin))
	if err != nil {
		panic(err)
	}
}

type HrSignPlugin struct{}

func (hs *HrSignPlugin) GetType() string {
	return PLUGIN_TYPE_HRSIGN
}
func (hs *HrSignPlugin) Execute(up *user.UserPlugin) (bool, error) {
	hrUid := up.Param("UserID")
	if hrUid == "" {
		return false, errors.New("no UserID in parameters")
	}

	real_url := fmt.Sprintf(FMT_HRSIGN_NOTICECHECK_URL, hrUid)
	resp, err := http.Get(real_url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	notice_resp := new(web.Response)
	err = json.Unmarshal(data, notice_resp)
	if err != nil {
		return false, err
	}

	flag, ok := notice_resp.Data.(bool)
	if !ok {
		return false, ERR_NOTICE_CHECK
	}

	// 消息内容
	// flag = true
	if flag {
		resp, err = http.Get(fmt.Sprintf(FMT_HRSIGN_SIGNLIST_URL, hrUid))
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		lsResp := new(hhsecret.ListSignResp)
		err = json.Unmarshal(data, lsResp)
		if err != nil {
			return false, err
		}

		var location = "未打卡"
		var signTime = "--"
		var tip = up.Param("tip")
		if len(lsResp.Data.Signs) > 0 {
			location = lsResp.Data.Signs[0].Location
			signTime = time.Unix(lsResp.Data.Signs[0].DateTime/1000, 0).Format("15:04")

			if tip == "" {
				tip = "下班打卡"
			}
		} else {
			if tip == "" {
				tip = "上班打卡"
			}
		}

		up.Values = []string{
			up.Param("name"),
			"微信打卡",
			location,
			signTime,
			tip,
		}
		up.Parameters["emphasis"] = []string{"5", "放大关键词", ""}
	}

	// if flag {
	// 	return true
	// 	log.Printf("should send a notice: %s!", ups.String())
	// } else {
	// 	log.Printf("can't send notice: %s!", ups.String())
	// }
	return flag, nil
}
func (hs *HrSignPlugin) GetTemplateMsgID() string {
	return "8U98v1g7PWLZ5p4jbWNSpY5dr-hhG5kVuMAUew4PHnY"
}

func (hs *HrSignPlugin) GetEmphasisID() string {
	return "5"
}

func (hs *HrSignPlugin) GetPage() string {
	return "/pages/hrsign/hrsign"
}
