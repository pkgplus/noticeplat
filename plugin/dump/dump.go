package dump

import (
	"github.com/xuebing1110/noticeplat/user"
	"log"
)

type DumpPlugin struct {
}

func (dp *DumpPlugin) GetType() string {
	return "dumpPlugin"
}

func (dp *DumpPlugin) Execute(ups *user.UserPluginSetting) error {
	log.Printf("%s\n", ups.String())
	return nil
}
