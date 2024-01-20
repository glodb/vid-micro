package controllers

import (
	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

type RolesController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u RolesController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u RolesController) GetCollectionName() basetypes.CollectionName {
	return "roles"
}

func (u RolesController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.Session{})

	role := models.Roles{
		Id:   1,
		Name: "admin",
		Slug: "admin",
	}

	u.Add(u.GetDBName(), u.GetCollectionName(), role, true)
	cache.GetInstance().HashMultiSet("auth_roles_1", map[string]interface{}{"name": "admin", "slug": "admin"})

	role = models.Roles{
		Id:   2,
		Name: "manager",
		Slug: "manager",
	}

	cache.GetInstance().HashMultiSet("auth_roles_2", map[string]interface{}{"name": "manager", "slug": "manager"})
	u.Add(u.GetDBName(), u.GetCollectionName(), role, true)

	role = models.Roles{
		Id:   20,
		Name: "user",
		Slug: "user",
	}

	cache.GetInstance().HashMultiSet("auth_roles_20", map[string]interface{}{"name": "user", "slug": "user"})
	u.Add(u.GetDBName(), u.GetCollectionName(), role, true)

	return nil
}

func (u *RolesController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u RolesController) RegisterApis() {
}
