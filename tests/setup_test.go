package tests

import (
	"testing"

	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/service"
	userimpl "github.com/LandcLi/LandcUserManager/internal/service_impl/user"
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
