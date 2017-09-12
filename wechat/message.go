package wechat

import (
	"fmt"

	"encoding/json"
)

type TemplateMsg struct {
	ToUserID        string                          `json:"touser"`
	TemplateID      string                          `json:"template_id"`
	FormID          string                          `json:"form_id"`
	Data            map[string]TemplateMsgDataValue `json:"data"`
	EmphasisKeyword string                          `json:"emphasis_keyword,omitempty"`
}

type TemplateMsgDataValue struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

func NewTemplateMsgData(values []string) map[string]TemplateMsgDataValue {
	data := make(map[string]TemplateMsgDataValue)
	for i, value := range values {
		data[fmt.Sprintf("keyword%d", i+1)] = TemplateMsgDataValue{Value: value}
	}

	return data
}

func NewTemplateMsg(userid, templateid, formid string, values []string) *TemplateMsg {
	msg := &TemplateMsg{
		ToUserID:   userid,
		TemplateID: templateid,
		FormID:     formid,
		Data:       NewTemplateMsgData(values),
	}

	data, _ := json.Marshal(msg)
	fmt.Printf("%s\n", data)

	return msg
}

func (tmsg *TemplateMsg) SetEmphasis(i int) {
	tmsg.EmphasisKeyword = fmt.Sprintf("keyword%d.DATA", i)
}