package storage

import (
	"github.com/xuebing1110/noticeplat/user"
	"github.com/xuebing1110/noticeplat/wechat"
)

type Storage interface {
	SaveSession(sess_3rd string, sessInfo *wechat.SessionResp) error
	QuerySession(sess_3rd string) (*wechat.SessionResp, error)

	UpsertUser(user user.User) error
	AddUser(user user.User) error
	Exist(uid string) bool
	AddEnergy(uid, energy string) error
	GetEnergyCount(uid string) int64
	PopEnergy(uid string) (string, error)

	AddUserPlugin(up *user.UserPlugin) error
	DelUserPlugin(unionid, pluginid string) error
	FetchTasks(curtime int64, handler func(*user.UserPlugin) error) error
}
