package application

import (
	"encoding/json"
	"mayfly-go/internal/sys/domain/entity"
	"mayfly-go/internal/sys/domain/repository"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/cache"
	"mayfly-go/pkg/logx"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/utils/jsonx"
	"strings"
)

const SysConfigKeyPrefix = "mayfly:sys:config:"

type Config interface {
	GetPageList(condition *entity.Config, pageParam *model.PageParam, toEntity any, orderBy ...string) *model.PageResult[any]

	Save(config *entity.Config)

	// GetConfig 获取指定key的配置信息, 不会返回nil, 若不存在则值都默认值即空字符串
	GetConfig(key string) *entity.Config
}

func newConfigApp(configRepo repository.Config) Config {
	return &configAppImpl{
		configRepo: configRepo,
	}
}

type configAppImpl struct {
	configRepo repository.Config
}

func (a *configAppImpl) GetPageList(condition *entity.Config, pageParam *model.PageParam, toEntity any, orderBy ...string) *model.PageResult[any] {
	return a.configRepo.GetPageList(condition, pageParam, toEntity)
}

func (a *configAppImpl) Save(config *entity.Config) {
	if config.Id == 0 {
		a.configRepo.Insert(config)
	} else {
		oldConfig := a.GetConfig(config.Key)
		if oldConfig.Permission != "all" {
			biz.IsTrue(strings.Contains(oldConfig.Permission, config.Modifier), "您无权修改该配置")
		}

		a.configRepo.Update(config)
	}
	cache.Del(SysConfigKeyPrefix + config.Key)
}

func (a *configAppImpl) GetConfig(key string) *entity.Config {
	config := &entity.Config{Key: key}
	// 优先从缓存中获取
	cacheStr := cache.GetStr(SysConfigKeyPrefix + key)
	if cacheStr != "" {
		json.Unmarshal([]byte(cacheStr), &config)
		return config
	}

	if err := a.configRepo.GetConfig(config, "Id", "Key", "Value", "Permission"); err != nil {
		logx.Warnf("不存在key = [%s] 的系统配置", key)
	} else {
		cache.SetStr(SysConfigKeyPrefix+key, jsonx.ToStr(config), -1)
	}
	return config
}
