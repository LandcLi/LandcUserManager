package tests

import (
	"os"
	"testing"
	"time"

	frameauth "github.com/LandcLi/landc-go/frame/pkg/auth"

	"github.com/LandcLi/LandcUserManager/dao"
	userimpl "github.com/LandcLi/LandcUserManager/internal/service_impl/user"
	"github.com/LandcLi/LandcUserManager/service"
)

var (
	mockUserDao          *MockUserDao
	mockLoginHistoryDao  *MockLoginHistoryDao
	mockThirdPartyDao    *MockThirdPartyUserDao
	mockTrustedDeviceDao *MockTrustedDeviceDao
	mockCommonService    *MockCommonService
)

func TestMain(m *testing.M) {
	// Register mock DAOs and services once (di.Provide does NOT override)
	mockUserDao = NewMockUserDao()
	mockLoginHistoryDao = NewMockLoginHistoryDao()
	mockThirdPartyDao = NewMockThirdPartyUserDao()
	mockTrustedDeviceDao = NewMockTrustedDeviceDao()
	mockCommonService = NewMockCommonService()

	dao.RegisterUserDao(mockUserDao)
	dao.RegisterLoginHistoryDao(mockLoginHistoryDao)
	dao.RegisterThirdPartyUserDao(mockThirdPartyDao)
	dao.RegisterTrustedDeviceDao(mockTrustedDeviceDao)
	service.RegisterCommonService(mockCommonService)

	// 初始化 JWT（测试密钥，长度 >= 32）
	frameauth.InitJWT(&frameauth.JWTConfig{
		Secret:     "test-secret-key-0123456789-abcdef-012345",
		ExpireTime: time.Hour,
		Issuer:     "landc-user-manager-test",
	})
	// 避免 crypto 包在测试打印开发密钥告警（32 hex 字符 → 16 字节 AES-128）
	os.Setenv("LUM_ENC_KEY", "0123456789abcdef0123456789abcdef")

	m.Run()
}

// reset clears all mock data before each test
func reset() {
	mockUserDao.Reset()
	mockLoginHistoryDao.Reset()
	mockThirdPartyDao.Reset()
	mockTrustedDeviceDao.Reset()
	mockCommonService.Reset()
}

func getUserService() service.UserService {
	return userimpl.NewUserServiceImpl()
}
