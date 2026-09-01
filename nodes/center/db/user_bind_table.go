package db

import (
	"fmt"

	"github.com/bychannel/cherry-server/internal/guid"
	cherryTime "github.com/cherry-game/cherry/extend/time"
)

// UserBindTable userid绑定第三方平台表
type UserBindTable struct {
	UserId    int64  `gorm:"column:user_id;primary_key;comment:'用户唯一id'" json:"userId"`
	SdkId     int32  `gorm:"column:sdk_id;comment:'sdk id'" json:"sdkId"`
	PackageId int32  `gorm:"column:package_id;comment:'平台包id'" json:"packageId"`
	OpenId    string `gorm:"column:open_id;comment:'平台帐号open_id'" json:"openId"`
	BindTime  int64  `gorm:"column:bind_time;comment:'绑定时间'" json:"bindTime"`
}

func (*UserBindTable) TableName() string {
	return "user_bind"
}

func GetUserId(packageId int32, openId string) (int64, bool) {
	cacheKey := fmt.Sprintf(userIdKey, packageId, openId)

	val, found := userIdCache.GetIfPresent(cacheKey)
	if found == false {
		return 0, false
	}

	return val.(int64), true
}

// BindUserId 绑定userId
func BindUserId(sdkId, packageId int32, openId string) (int64, bool) {
	// TODO 根据 platformType的配置要求，决定查询UID的方式：
	// 条件1: platformType + openId查询，是否存在uid
	// 条件2: packageId + openId查询，是否存在uid

	userId, ok := GetUserId(packageId, openId)
	if ok {
		return userId, true
	}

	userBind := &UserBindTable{
		UserId:    guid.Next(),
		SdkId:     sdkId,
		PackageId: packageId,
		OpenId:    openId,
		BindTime:  cherryTime.Now().ToMillisecond(),
	}

	cacheKey := fmt.Sprintf(userIdKey, packageId, openId)
	userIdCache.Put(cacheKey, userBind.UserId)

	return userBind.UserId, true
}
