package serverlib

import (
	"fmt"

	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"gorm.io/gorm"

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

// emailIndexName 是 users.email 上的索引名（GORM 按「idx_<表>_<列>」生成）。
const emailIndexName = "idx_users_email"

// EnsureSchema 按配置决定是否自动建表/补列。
//
// 受框架配置 database.auto_migrate 控制：
//   - SQLite / 开发测试环境建议开启，开箱即用；
//   - MySQL 生产环境建议关闭，由 sqls/init.sql 或版本化迁移接管。
func EnsureSchema() error {
	// 必须在 AutoMigrate 之前：历史版本给 email 建的是唯一索引，
	// 而 AutoMigrate 不会把已有的唯一索引改成普通索引（它只补缺）。
	if err := dropLegacyUniqueEmailIndex(framedb.GetDB()); err != nil {
		return fmt.Errorf("serverlib: drop legacy email unique index: %w", err)
	}

	if err := framedb.AutoMigrateIfEnabled(Models()...); err != nil {
		return fmt.Errorf("serverlib: migrate schema: %w", err)
	}
	return nil
}

// dropLegacyUniqueEmailIndex 删除历史遗留的「email 唯一索引」。
//
// 背景：早期 schema 把 email 定义为 NOT NULL + UNIQUE，但邮箱其实是可选的
// （用户名/手机号注册、第三方登录都可能没有邮箱）。SQL 里空串是确定值，
// 于是「不填邮箱的用户」全表只能存在一个 —— 第二个用户名注册直接报
// UNIQUE constraint failed。这不是理论问题，是第二个人注册就复现的阻断级缺陷。
//
// 现在改为普通索引，非空邮箱的唯一性由应用层在注册/改绑时校验。
// 这里做的是幂等清理：索引不存在时按方言静默跳过。
func dropLegacyUniqueEmailIndex(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	var stmt string
	switch db.Dialector.Name() {
	case "sqlite":
		// IF EXISTS 保证幂等；删掉后 AutoMigrate 会按新 tag 重建为普通索引。
		stmt = fmt.Sprintf("DROP INDEX IF EXISTS %s", emailIndexName)
	case "postgres":
		stmt = fmt.Sprintf("DROP INDEX IF EXISTS %s", emailIndexName)
	case "mysql":
		// MySQL 没有 DROP INDEX IF EXISTS，不存在时报错即可忽略。
		stmt = fmt.Sprintf("ALTER TABLE users DROP INDEX %s", emailIndexName)
	default:
		return nil
	}

	if err := db.Exec(stmt).Error; err != nil {
		// 索引本来就不存在是最常见的情况，不算失败。
		return nil
	}
	return nil
}
