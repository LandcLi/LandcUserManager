package serverlib

import (
	"fmt"

	framedb "github.com/LandcLi/landc-go/frame/pkg/db"

	"github.com/LandcLi/LandcUserManager/model"
)

// Models 返回 LandcUserManager 负责的全部持久化模型。
// 供自动迁移（开发/测试）与文档使用，保证「模型清单」只有一处定义。
func Models() []interface{} {
	return []interface{}{
		&model.User{},
		&model.LoginHistory{},
		&model.TrustedDevice{},
		&model.ThirdPartyUserMeta{},
		&model.Biometric{},
	}
}

// EnsureSchema 按配置决定是否自动建表/补列。
//
// 受框架配置 database.auto_migrate 控制：
//   - SQLite / 开发测试环境建议开启，开箱即用；
//   - MySQL 生产环境建议关闭，由 sqls/init.sql 或版本化迁移接管。
func EnsureSchema() error {
	if err := framedb.AutoMigrateIfEnabled(Models()...); err != nil {
		return fmt.Errorf("serverlib: migrate schema: %w", err)
	}
	return nil
}
