package modules

import (
	_ "react-go-admin/app/internal/modules/apis"
	_ "react-go-admin/app/internal/modules/auditlog"
	_ "react-go-admin/app/internal/modules/base"
	_ "react-go-admin/app/internal/modules/install"
	_ "react-go-admin/app/internal/modules/roles"
	_ "react-go-admin/app/internal/modules/systemsettings"
	_ "react-go-admin/app/internal/modules/upload"
	_ "react-go-admin/app/internal/modules/users"
)

// Register exists to make the side-effect imports explicit at the call site.
func Register() {}
