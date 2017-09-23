package user

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xuebing1110/noticeplat/plugin/cron"
)

type UserPlugin struct {
	UserID   string `json:"userID"`
	PluginID string `json:"pluginID"`
	*UserPluginSetting
}

type UserPluginSetting struct {
	Desc        string            `json:"desc"`
	CronSetting *cron.CronSetting `json:"cronSetting"`

	Parameters map[string]string `json:"parameters"`
	Values     []string          `json:"values"`
	PluginType string            `json:"pluginType"`
	CreateTime int64             `json:"createTime"`
	Disable    bool              `json:"disable"`
}

// type Parameter struct {
// 	Label     string `json:"label,omitempty"`
// 	LabelDesc string `json:"labelDesc,omitempty"`
// 	Value     string `json:"value,omitempty"`
// 	ValueDesc string `json:"valueDesc,omitempty"`
// }

func NewUserPlugin(uid, pluginid string, setting []byte) (*UserPlugin, error) {
	ups, err := NewUserPluginSetting(setting)
	if err != nil {
		return nil, err
	}

	return &UserPlugin{
		UserID:            uid,
		PluginID:          pluginid,
		UserPluginSetting: ups,
	}, nil
}

func NewUserPluginSetting(data []byte) (usetting *UserPluginSetting, err error) {
	usetting = new(UserPluginSetting)
	err = json.Unmarshal(data, usetting)
	if err != nil {
		return nil, err
	}

	if usetting.CronSetting == nil {
		return nil, fmt.Errorf("cronSetting parsed failed")
	}

	if usetting.CreateTime == 0 {
		usetting.CreateTime = time.Now().UnixNano()
	}

	err = usetting.CronSetting.Init()
	return
}

func (ups *UserPluginSetting) Param(key string) string {
	values, found := ups.Parameters[key]
	if found {
		return values
	} else {
		return ""
	}
}

func (ups *UserPluginSetting) String() string {
	bytes, _ := json.Marshal(ups)
	return string(bytes)
}
