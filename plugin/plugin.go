package plugin

import (
	"errors"
	"sync"

	"github.com/xuebing1110/noticeplat/user"
)

type plugins struct {
	sync.Mutex
	detail map[string]Plugin
}

var (
	defaultPlugins *plugins
)

func init() {
	defaultPlugins = &plugins{
		detail: make(map[string]Plugin),
	}
}

type Plugin interface {
	GetType() string
	Execute(*user.UserPlugin) (bool, error)
	GetTemplateMsgID() string
	GetEmphasisID() string
	GetPage() string
}

func Registe(plugintype string, plugin Plugin) error {
	defaultPlugins.Lock()
	defer defaultPlugins.Unlock()

	_, found := defaultPlugins.detail[plugintype]
	if found {
		return errors.New("plugin is registed")
	}

	defaultPlugins.detail[plugintype] = plugin
	return nil
}

func GetPlugin(plugintype string) (Plugin, error) {
	p, found := defaultPlugins.detail[plugintype]
	if !found {
		return nil, errors.New("plugin not found")
	}

	return p, nil
}
