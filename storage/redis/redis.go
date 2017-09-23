package redis

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	// "github.com/xuebing1110/noticeplat/plugin"
	// "github.com/xuebing1110/noticeplat/plugin/cron"
	"github.com/xuebing1110/noticeplat/user"
	"github.com/xuebing1110/noticeplat/wechat"
)

const (
	SESS_PREFIX        = "sess."
	USER_PREFIX        = "user."
	ENERGY_PREFIX      = "energy."
	USERPLUGINS_PREFIX = "userplugins."
	TASKS_SORTSET      = "tasks"
	EXEC_MAXINTERVAL   = 20
	ASYNC_MAXTASKS     = 1000
)

var (
	Random                 = rand.New(rand.NewSource(time.Now().Unix()))
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

func (rs *RedisStorage) ListUserPlugins(uid string) (ups []*user.UserPlugin, err error) {
	// userplugins
	ret := rs.HGetAll(USERPLUGINS_PREFIX + uid)
	if ret.Err() != nil {
		return ups, ret.Err()
	}

	ups = make([]*user.UserPlugin, 0, len(ret.Val()))
	for pluginid, up_bytes := range ret.Val() {
		up, err := user.NewUserPlugin(uid, pluginid, []byte(up_bytes))
		if err != nil {
			log.Printf("parse userPlugin failed err:%v, body:%s", err, up_bytes)
		} else {
			ups = append(ups, up)
		}
	}

	return ups, nil
}

func (rs *RedisStorage) AddUserPlugin(up *user.UserPlugin) error {
	// runtime
	cronSetting := up.CronSetting
	runtime := cronSetting.NextRunTime(time.Now().Truncate(time.Minute))
	if runtime.IsZero() {
		return errors.New(ERROR_USERTASK_EXPIRED.Error() + ":" + cronSetting.String())
	}

	// userplugins
	ret := rs.HSet(USERPLUGINS_PREFIX+up.UserID, up.PluginID, up.String())
	if ret.Err() != nil {
		return ret.Err()
	}

	// tasks
	zret := rs.ZAdd(TASKS_SORTSET, redis.Z{
		float64(runtime.Unix()),
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

func (rs *RedisStorage) GetUserPlugin(uid, pluginid string) (*user.UserPlugin, error) {
	ret := rs.HGet(USERPLUGINS_PREFIX+uid, pluginid)
	if ret.Err() != nil {
		return nil, ret.Err()
	}

	return user.NewUserPlugin(uid, pluginid, []byte(ret.Val()))
}

func (rs *RedisStorage) FetchTasks(curtime int64, handler func(*user.UserPlugin) error) error {
	// log.Printf("prepare to get task: 0-%d", curtime)
	ret := rs.ZRevRangeByScoreWithScores(
		TASKS_SORTSET,
		redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%d", curtime),
		})
	retZs, err := ret.Result()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var asyncLimit = make(chan bool, ASYNC_MAXTASKS)
	for _, retZ := range retZs {
		// log.Printf("get task:%+v", retZ.Member)
		asyncLimit <- true
		wg.Add(1)
		go func(retZ redis.Z) {
			defer func() {
				wg.Done()
				<-asyncLimit
			}()

			err := rs.oneTask(curtime, handler, retZ)
			if err != nil {
				log.Printf("handler %s err:%v\n", retZ.Member.(string), err)
			}
		}(retZ)
	}

	wg.Wait()

	return nil
}

func (rs *RedisStorage) oneTask(curtime int64, handler func(*user.UserPlugin) error, retZ redis.Z) error {
	uid_pid_str := retZ.Member.(string)
	uid_pid := strings.SplitN(uid_pid_str, ":", 2)
	setting_ret := rs.HGet(USERPLUGINS_PREFIX+uid_pid[0], uid_pid[1])
	if setting_ret.Err() != nil {
		log.Printf("hget %s %s err:%v\n", USERPLUGINS_PREFIX+uid_pid[0], uid_pid[1], setting_ret.Err())
		if setting_ret.Err() == redis.Nil {
			ret2 := rs.ZRem(TASKS_SORTSET, uid_pid_str)
			log.Printf("zrem %s %s, result:%v\n", TASKS_SORTSET, uid_pid, ret2.Err())
		}
		return setting_ret.Err()
	}

	pluginSetting, err := setting_ret.Result()
	if err != nil {
		log.Printf("hget %s %s result err:%v\n", USERPLUGINS_PREFIX+uid_pid[0], uid_pid[1], setting_ret.Err())
		if setting_ret.Err() == redis.Nil {
			ret2 := rs.ZRem(TASKS_SORTSET, uid_pid_str)
			log.Printf("zrem %s %s, result:%v\n", TASKS_SORTSET, uid_pid[1], ret2.Err())
		}
		return err
	}

	// log.Printf("pluginSetting:  %s\n", pluginSetting)
	userPlugin, err := user.NewUserPlugin(uid_pid[0], uid_pid[1], []byte(pluginSetting))
	if err != nil {
		log.Printf("parse setting %s err:%s\n", setting_ret.Val(), err)
		if setting_ret.Err() == redis.Nil {
			ret2 := rs.ZRem(TASKS_SORTSET, uid_pid_str)
			log.Printf("zrem %s %s, result:%v\n", TASKS_SORTSET, uid_pid, ret2.Err())
		}
		return err
	}

	time.Sleep(time.Duration(Random.Int()%EXEC_MAXINTERVAL) * time.Second)
	log.Printf("deal %s user plugin:%s\n", uid_pid_str, userPlugin.String())
	if !userPlugin.Disable {
		err = handler(userPlugin)
		if err != nil {
			return fmt.Errorf("handle %s err:%v\n", setting_ret.Val(), err)
		}
	}

	var next_runtime = userPlugin.CronSetting.NextRunTime(time.Unix(curtime, 0))
	ret2 := rs.ZAdd(TASKS_SORTSET, redis.Z{
		float64(next_runtime.Unix()),
		uid_pid_str,
	})

	return ret2.Err()
}
