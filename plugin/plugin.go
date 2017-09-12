package plugin

import (
	"github.com/xuebing1110/noticeplat/user"
)

type Plugin interface {
	GetType() string
	Execute(*user.UserPluginSetting) error
	GetTemplateMsgID() string
}
