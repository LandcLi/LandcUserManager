package internal

import (
	_ "github.com/LandcLi/LandcUserManager/internal/dao_impl/login_history"
	_ "github.com/LandcLi/LandcUserManager/internal/dao_impl/third_party_user"
	_ "github.com/LandcLi/LandcUserManager/internal/dao_impl/trusted_device"
	_ "github.com/LandcLi/LandcUserManager/internal/dao_impl/user"

	_ "github.com/LandcLi/LandcUserManager/internal/service_impl/third_party"
	_ "github.com/LandcLi/LandcUserManager/internal/service_impl/user"

	_ "github.com/LandcLi/LandcUserManager/internal/controller/user"
)
