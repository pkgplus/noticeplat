package redis

import (
	// "log"
	"time"

	"github.com/xuebing1110/noticeplat/plugin/cron"
	"github.com/xuebing1110/noticeplat/plugin/hrsign"
	"github.com/xuebing1110/noticeplat/user"
)

import (
	"testing"
)

func TestRedisStorage(t *testing.T) {
	u := user.NewUser(map[string]interface{}{
		user.USER_FIELD_UNIONID: "xxxxxxxxxxx",
		user.USER_FIELD_SUBTIME: time.Now().Unix(),
	})

	// add user
	err := Client.AddUser(u)
	if err != nil {
		t.Fatal(err)
	}

	if !Client.Exist(u.ID()) {
		t.Fatal("call Exist failed: can't found " + u.ID())
	}

	// enery
	err = Client.AddEnergy(u.ID(), "1111111111")
	if err != nil {
		t.Fatal(err)
	}

	// enery
	eCount := Client.GetEnergyCount(u.ID())
	if eCount != 1 {
		t.Fatalf("energy expect == 1,but get %d!", eCount)
	}

	endrgyContent, err := Client.PopEnergy(u.ID())
	if err != nil {
		t.Fatal(err)
	}
	if endrgyContent != "1111111111" {
		t.Fatalf("pop energy: expect 1111111111,but get %s", endrgyContent)
	}

	// user plugin
	up := &user.UserPlugin{
		UserID:   "xxxxxxxxxxx",
		PluginID: "1",
		Setting: &user.UserPluginSetting{
			CronSetting: &cron.Setting{
				First:     1504850400,
				Intervals: []string{"@every 10s"},
			},
			PluginType: "HrSign",
			Parameters: map[string]string{
				"key": "value",
			},
		},
	}
	err = Client.AddUserPlugin(up)
	if err != nil {
		t.Fatal(err)
	}

	curtime := time.Now().Unix()
	err = Client.FetchTasks(curtime,
		func(ups *user.UserPlugin) error {
			if ups.Setting.CronSetting.First != 1504850400 ||
				ups.Setting.CronSetting.Intervals[0] != "@every 10s" ||
				ups.Setting.PluginType != "HrSign" {
				t.Fatalf("get unkown user plugin setting:%s", ups.String())
			} else {
				sign := &hrsign.HrSignPlugin{HrUserID: "01462834"}
				return sign.Execute(ups)
			}
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	err = Client.DelUserPlugin(up.UnionID, up.PluginID)
	if err != nil {
		t.Fatal(err)
	}
}
