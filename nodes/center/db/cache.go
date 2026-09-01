package db

import (
	"github.com/goburrow/cache"
	"time"
)

var (

	// userid缓存 key:userIdKey, value:userid
	userIdCache = cache.New(
		cache.WithMaximumSize(65535),
		cache.WithExpireAfterAccess(120*time.Minute),
	)

	// 开发帐号缓存 key:accountName, value:DevAccountTable
	devAccountCache = cache.New(
		cache.WithMaximumSize(65535),
		cache.WithExpireAfterAccess(60*time.Minute),
	)
)

// cache key
const (
	userIdKey = "userid.%d.%s" // packageId,openId
)
