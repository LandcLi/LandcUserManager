package tests

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/LandcLi/LandcUserManager/model"
)

// ============== MockUserDao ==============

type MockUserDao struct {
	mu    sync.Mutex
	users map[uint]*model.User
	byUN  map[string]*model.User // username
	byEM  map[string]*model.User // email
	byPH  map[string]*model.User // phone
}

func NewMockUserDao() *MockUserDao {
	return &MockUserDao{
		users: make(map[uint]*model.User),
		byUN:  make(map[string]*model.User),
		byEM:  make(map[string]*model.User),
		byPH:  make(map[string]*model.User),
	}
}

func (m *MockUserDao) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.users)
	clear(m.byUN)
	clear(m.byEM)
	clear(m.byPH)
}

func (m *MockUserDao) Create(ctx context.Context, user *model.User) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// auto-assign ID
	user.ID = uint(len(m.users) + 1)
	if user.Status == 0 {
		user.Status = 1 // matches gorm:"default:1" in model
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	if user.Username != "" {
		m.byUN[user.Username] = user
	}
	if user.Email != "" {
		m.byEM[user.Email] = user
	}
	if user.Phone != "" {
		m.byPH[user.Phone] = user
	}
	return user, nil
}

func (m *MockUserDao) FindByID(ctx context.Context, id uint) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *MockUserDao) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byUN[username]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *MockUserDao) FindByUsernameFromDB(ctx context.Context, username string) (*model.User, error) {
	return m.FindByUsername(ctx, username)
}

func (m *MockUserDao) FindByEmailFromDB(ctx context.Context, email string) (*model.User, error) {
	return m.FindByEmail(ctx, email)
}

func (m *MockUserDao) FindByPhoneFromDB(ctx context.Context, phone string) (*model.User, error) {
	return m.FindByPhone(ctx, phone)
}

func (m *MockUserDao) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byEM[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *MockUserDao) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byPH[phone]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *MockUserDao) Update(ctx context.Context, user *model.User) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.users[user.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	if existing.Username != "" && existing.Username != user.Username {
		delete(m.byUN, existing.Username)
	}
	if existing.Email != "" && existing.Email != user.Email {
		delete(m.byEM, existing.Email)
	}
	if existing.Phone != "" && existing.Phone != user.Phone {
		delete(m.byPH, existing.Phone)
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	if user.Username != "" {
		m.byUN[user.Username] = user
	}
	if user.Email != "" {
		m.byEM[user.Email] = user
	}
	if user.Phone != "" {
		m.byPH[user.Phone] = user
	}
	return user, nil
}

func (m *MockUserDao) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return errors.New("not found")
	}
	if u.Username != "" {
		delete(m.byUN, u.Username)
	}
	if u.Email != "" {
		delete(m.byEM, u.Email)
	}
	if u.Phone != "" {
		delete(m.byPH, u.Phone)
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserDao) ClearUserCache(ctx context.Context, user *model.User) error {
	return nil
}

func (m *MockUserDao) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var users []*model.User
	for _, u := range m.users {
		users = append(users, u)
	}
	total := int64(len(users))
	return users, total, nil
}

// Helper to add a pre-existing user
func (m *MockUserDao) AddUser(user *model.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.ID = uint(len(m.users) + 1)
	m.users[user.ID] = user
	if user.Username != "" {
		m.byUN[user.Username] = user
	}
	if user.Email != "" {
		m.byEM[user.Email] = user
	}
	if user.Phone != "" {
		m.byPH[user.Phone] = user
	}
}

// ============== MockLoginHistoryDao ==============

type MockLoginHistoryDao struct {
	mu      sync.Mutex
	records []*model.LoginHistory
}

func NewMockLoginHistoryDao() *MockLoginHistoryDao {
	return &MockLoginHistoryDao{
		records: make([]*model.LoginHistory, 0),
	}
}

func (m *MockLoginHistoryDao) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = make([]*model.LoginHistory, 0)
}

func (m *MockLoginHistoryDao) Create(ctx context.Context, lh *model.LoginHistory) (*model.LoginHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	lh.ID = uint(len(m.records) + 1)
	lh.CreatedAt = time.Now()
	m.records = append(m.records, lh)
	return lh, nil
}

func (m *MockLoginHistoryDao) FindByID(ctx context.Context, id uint) (*model.LoginHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.records {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockLoginHistoryDao) FindByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.LoginHistory, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*model.LoginHistory
	for _, r := range m.records {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	total := int64(len(result))
	return result, total, nil
}

func (m *MockLoginHistoryDao) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.records {
		if r.ID == id {
			m.records = append(m.records[:i], m.records[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *MockLoginHistoryDao) DeleteByUserID(ctx context.Context, userID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var keep []*model.LoginHistory
	for _, r := range m.records {
		if r.UserID != userID {
			keep = append(keep, r)
		}
	}
	m.records = keep
	return nil
}

func (m *MockLoginHistoryDao) DeleteOldRecords(ctx context.Context, days int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var keep []*model.LoginHistory
	for _, r := range m.records {
		if r.CreatedAt.After(cutoff) {
			keep = append(keep, r)
		}
	}
	m.records = keep
	return nil
}

// ============== MockThirdPartyUserDao ==============

type MockThirdPartyUserDao struct {
	mu    sync.Mutex
	metas []*model.ThirdPartyUserMeta
}

func NewMockThirdPartyUserDao() *MockThirdPartyUserDao {
	return &MockThirdPartyUserDao{
		metas: make([]*model.ThirdPartyUserMeta, 0),
	}
}

func (m *MockThirdPartyUserDao) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metas = make([]*model.ThirdPartyUserMeta, 0)
}

func (m *MockThirdPartyUserDao) Create(ctx context.Context, meta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta.ID = uint(len(m.metas) + 1)
	meta.CreatedAt = time.Now()
	meta.UpdatedAt = time.Now()
	m.metas = append(m.metas, meta)
	return meta, nil
}

func (m *MockThirdPartyUserDao) FindByID(ctx context.Context, id uint) (*model.ThirdPartyUserMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, meta := range m.metas {
		if meta.ID == id {
			return meta, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockThirdPartyUserDao) FindByOpenIDAndProvider(ctx context.Context, openID string, provider string) (*model.ThirdPartyUserMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, meta := range m.metas {
		if meta.OpenID == openID && meta.Provider == provider {
			return meta, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockThirdPartyUserDao) FindByUnionID(ctx context.Context, unionID string) (*model.ThirdPartyUserMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, meta := range m.metas {
		if meta.UnionID == unionID {
			return meta, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockThirdPartyUserDao) FindByUserID(ctx context.Context, userID uint) ([]*model.ThirdPartyUserMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*model.ThirdPartyUserMeta
	for _, meta := range m.metas {
		if meta.UserID == userID {
			result = append(result, meta)
		}
	}
	return result, nil
}

func (m *MockThirdPartyUserDao) Update(ctx context.Context, meta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, existing := range m.metas {
		if existing.ID == meta.ID {
			meta.UpdatedAt = time.Now()
			m.metas[i] = meta
			return meta, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockThirdPartyUserDao) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, meta := range m.metas {
		if meta.ID == id {
			m.metas = append(m.metas[:i], m.metas[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *MockThirdPartyUserDao) DeleteByUserID(ctx context.Context, userID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var keep []*model.ThirdPartyUserMeta
	for _, meta := range m.metas {
		if meta.UserID != userID {
			keep = append(keep, meta)
		}
	}
	m.metas = keep
	return nil
}

// ============== MockTrustedDeviceDao ==============

type MockTrustedDeviceDao struct {
	mu      sync.Mutex
	devices []*model.TrustedDevice
}

func NewMockTrustedDeviceDao() *MockTrustedDeviceDao {
	return &MockTrustedDeviceDao{
		devices: make([]*model.TrustedDevice, 0),
	}
}

func (m *MockTrustedDeviceDao) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices = make([]*model.TrustedDevice, 0)
}

func (m *MockTrustedDeviceDao) Create(ctx context.Context, d *model.TrustedDevice) (*model.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d.ID = uint(len(m.devices) + 1)
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	m.devices = append(m.devices, d)
	return d, nil
}

func (m *MockTrustedDeviceDao) FindByID(ctx context.Context, id uint) (*model.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.devices {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockTrustedDeviceDao) FindByUserIDAndFingerprint(ctx context.Context, userID uint, fingerprint string) (*model.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.devices {
		if d.UserID == userID && d.DeviceFingerprint == fingerprint {
			return d, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockTrustedDeviceDao) FindByUserID(ctx context.Context, userID uint) ([]*model.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*model.TrustedDevice
	for _, d := range m.devices {
		if d.UserID == userID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MockTrustedDeviceDao) Update(ctx context.Context, d *model.TrustedDevice) (*model.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, existing := range m.devices {
		if existing.ID == d.ID {
			d.UpdatedAt = time.Now()
			m.devices[i] = d
			return d, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockTrustedDeviceDao) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, d := range m.devices {
		if d.ID == id {
			m.devices = append(m.devices[:i], m.devices[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *MockTrustedDeviceDao) DeleteByUserID(ctx context.Context, userID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var keep []*model.TrustedDevice
	for _, d := range m.devices {
		if d.UserID != userID {
			keep = append(keep, d)
		}
	}
	m.devices = keep
	return nil
}
