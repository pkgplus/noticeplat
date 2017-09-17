package user

import (
	"encoding/json"

	"github.com/xuebing1110/noticeplat/plugin/cron"
)

type UserPlugin struct {
	UserID   string
	PluginID string
	*UserPluginSetting
}

type UserPluginSetting struct {
	CronSetting *cron.CronSetting

	Parameters map[string][]string
	Values     []string
	PluginType string
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

	err = usetting.CronSetting.Init()
	return
}

func (ups *UserPluginSetting) Param(key string) string {
	return ups.Parameters[key][0]
}

func (ups *UserPluginSetting) String() string {
	bytes, _ := json.Marshal(ups)
	return string(bytes)
}
