package user

import (
	"time"
)

const (
	USER_FIELD_UNIONID  = "unionid"
	USER_FIELD_OPENID   = "openid"
	USER_FIELD_NICKNAME = "nickName"
	USER_FIELD_GENDER   = "gender"
	USER_FIELD_PROVINCE = "province"
	USER_FIELD_CITY     = "city"
	USER_FIELD_COUNTRY  = "country"
	USER_FIELD_SUBTIME  = "subTime"
)

type User map[string]interface{}

func NewUser(u map[string]interface{}) User {
	u[USER_FIELD_SUBTIME] = time.Now().Unix()
	return User(u)
}

func (u User) Get(field string) interface{} {
	umap := map[string]interface{}(u)
	return umap[field]
}

func (u User) GetString(field string) string {
	umap := map[string]interface{}(u)

	v, found := umap[field]
	if !found {
		return ""
	} else {
		return v.(string)
	}
}

func (u User) ID() string {
	return u.GetString(USER_FIELD_UNIONID)
}
