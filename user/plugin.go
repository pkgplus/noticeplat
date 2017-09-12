package user

import (
	"encoding/json"

	"github.com/xuebing1110/noticeplat/plugin/cron"
)

type UserPlugin struct {
	UserID   string
	PluginID string
	Setting  *UserPluginSetting
}

type UserPluginSetting struct {
	CronSetting *cron.Setting

	Parameters map[string]string
	PluginType string
}

func NewUserPluginSetting(data []byte) (usetting *UserPluginSetting, err error) {
	usetting = new(UserPluginSetting)
	err = json.Unmarshal(data, usetting)
	if err != nil {
		return nil, err
	}
	return
}

func (ups *UserPluginSetting) String() string {
	bytes, _ := json.Marshal(ups)
	return string(bytes)
}
