package redis

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	// "github.com/xuebing1110/noticeplat/plugin"
	"github.com/xuebing1110/noticeplat/plugin/cron"
	"github.com/xuebing1110/noticeplat/user"
	"github.com/xuebing1110/noticeplat/wechat"
)

const (
	SESS_PREFIX        = "sess."
	USER_PREFIX        = "user."
	ENERGY_PREFIX      = "energy."
	USERPLUGINS_PREFIX = "userplugins."
	TASKS_SORTSET      = "tasks"
)

var (
	SEVEN_DAY              = 168 * time.Hour
	Client                 *RedisStorage
	ERROR_USERTASK_EXPIRED = errors.New("taskHasExpired")
)

type RedisStorage struct {
	*redis.Client
}

func init() {
	// RedisClient
	addr := os.Getenv("REDIS_ADDR")
	passwd := os.Getenv("REDIS_PASSWD")
	if addr == "" {
		addr = "localhost:6379"
	}
	rc := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       0,
	})

	// RedisStorage
	Client = &RedisStorage{
		Client: rc,
	}
}

func (rs *RedisStorage) SaveSession(sess_3rd string, sessInfo *wechat.SessionResp) error {
	ret := rs.HMSet(SESS_PREFIX+sess_3rd, sessInfo.Convert2Map())
	// if ret.Err() != nil {
	// 	return ret.Err()
	// }
	return ret.Err()

	// expRet := rs.ExpireAt(SESS_PREFIX+sess_3rd, time.Now().Add(SEVEN_DAY))
	// return expRet.Err()
}

func (rs *RedisStorage) QuerySession(sess_3rd string) (*wechat.SessionResp, error) {
	ret := rs.HGetAll(SESS_PREFIX + sess_3rd)
	if ret.Err() != nil {
		return nil, ret.Err()
	}
	return wechat.NewSessionResp(ret.Val())
}

func (rs *RedisStorage) UpsertUser(user user.User) error {
	ret := rs.HMSet(USER_PREFIX+user.ID(), map[string]interface{}(user))
	return ret.Err()
}

func (rs *RedisStorage) AddUser(user user.User) error {
	return rs.UpsertUser(user)
}

func (rs *RedisStorage) Exist(uid string) bool {
	ret := rs.HGet(USER_PREFIX+uid, user.USER_FIELD_SUBTIME)
	if ret.Err() != nil {
		return false
	}

	_, err := ret.Int64()
	if err != nil {
		return false
	}

	return true
}

func (rs *RedisStorage) AddEnergy(uid, energy string) error {
	ret := rs.RPush(ENERGY_PREFIX+uid, fmt.Sprintf("%s,%d", energy, time.Now().Unix()))
	if ret.Err() != nil {
		return ret.Err()
	}

	rs.ExpireAt(ENERGY_PREFIX+uid, time.Now().Add(SEVEN_DAY))
	return nil
}

func (rs *RedisStorage) GetEnergyCount(uid string) int64 {
	ret := rs.LLen(ENERGY_PREFIX + uid)
	return ret.Val()
}

func (rs *RedisStorage) PopEnergy(uid string) (string, error) {
	var curtime int64 = time.Now().Unix()
	for {
		energy_ret, err := rs.popOneEnergy(uid)
		if err != nil {
			return "", err
		}
		energy_info := strings.SplitN(energy_ret, ",", 2)
		if len(energy_info) != 2 {
			return "", errors.New("text")
		}

		pushtime, err := strconv.Atoi(energy_info[1])
		if err != nil {
			log.Printf("convert to time failed:%s", err)
			continue
		}

		if curtime-int64(pushtime) < 604000 {
			return energy_info[0], nil
		}
	}

	return "", nil
}

func (rs *RedisStorage) popOneEnergy(uid string) (string, error) {
	ret := rs.LPop(ENERGY_PREFIX + uid)
	return ret.Result()
}

func (rs *RedisStorage) ExpireEnergy(uid string) error {
	ret := rs.ExpireAt(ENERGY_PREFIX+uid, time.Now().Add(SEVEN_DAY))
	return ret.Err()
}

func (rs *RedisStorage) AddUserPlugin(up *user.UserPlugin) error {
	// runtime
	cronSetting := up.CronSetting
	runtime := cronSetting.First
	if runtime == 0 {
		return errors.New(ERROR_USERTASK_EXPIRED.Error() + ":" + cronSetting.String())
	}

	// userplugins
	ret := rs.HSet(USERPLUGINS_PREFIX+up.UserID, up.PluginID, up.String())
	if ret.Err() != nil {
		return ret.Err()
	}

	// tasks
	zret := rs.ZAdd(TASKS_SORTSET, redis.Z{
		float64(runtime),
		up.UserID + ":" + up.PluginID,
	})
	return zret.Err()
}

func (rs *RedisStorage) DelUserPlugin(uid, pluginid string) error {
	ret := rs.HDel(USERPLUGINS_PREFIX+uid, pluginid)
	if ret.Err() != nil {
		return ret.Err()
	}

	ret = rs.ZRem(TASKS_SORTSET, uid+":"+pluginid)
	return ret.Err()
}

func (rs *RedisStorage) FetchTasks(curtime int64, handler func(*user.UserPlugin) error) error {
	ret := rs.ZRangeWithScores(TASKS_SORTSET, 0, curtime)
	retZs, err := ret.Result()
	if err != nil {
		return err
	}

	for _, retZ := range retZs {
		uid_pid := strings.SplitN(retZ.Member.(string), ":", 2)
		setting_ret := rs.HGet(USERPLUGINS_PREFIX+uid_pid[0], uid_pid[1])
		if setting_ret.Err() != nil {
			log.Printf("hget %s %s err:%v\n", USERPLUGINS_PREFIX+uid_pid[0], uid_pid[1], setting_ret.Err())
			continue
		}

		pluginSetting, err := setting_ret.Result()
		if err != nil {
			log.Printf("hget %s %s result err:%v\n", USERPLUGINS_PREFIX+uid_pid[0], uid_pid[1], setting_ret.Err())
			continue
		}

		// log.Printf("pluginSetting:  %s\n", pluginSetting)
		userPlugin, err := user.NewUserPlugin(uid_pid[0], uid_pid[1], []byte(pluginSetting))
		if err != nil {
			log.Printf("parse setting %s err:%s\n", setting_ret.String(), err)
			continue
		}

		err = handler(userPlugin)
		if err != nil {
			log.Printf("handle %s err:%v\n", setting_ret.String(), err)
			continue
		}

		next_runtime := cron.GetNextRunTime(setting_ret.String())
		rs.ZAdd(TASKS_SORTSET, redis.Z{
			float64(next_runtime),
			uid_pid,
		})

	}

	return nil
}
