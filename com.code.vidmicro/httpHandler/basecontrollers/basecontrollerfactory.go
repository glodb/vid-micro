package basecontrollers

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/app/controllers"
	"com.code.vidmicro/com.code.vidmicro/app/validators"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

var instance *controllersObject
var once sync.Once

// Controllers struct
type controllersObject struct {
	controllers map[string]baseinterfaces.Controller
}

// Singleton. Returns a single object of Factory
func GetInstance() *controllersObject {
	// var instance
	once.Do(func() {
		instance = &controllersObject{}
		instance.controllers = make(map[string]baseinterfaces.Controller)
	})
	return instance
}

// createController is a factory to return the appropriate controller
func (c *controllersObject) GetController(controllerType string) (baseinterfaces.Controller, error) {
	if _, ok := c.controllers[controllerType]; ok {
		return c.controllers[controllerType], nil
	} else {
		c.registerControllers(controllerType, false)
		return c.controllers[controllerType], nil
	}
}

/**
*To all developers are future me,
*Although this is lazy flyweight factory it doesn't work as lazy factory for web server
*It will register all the controllers defined in the config for web but it will still be flyweight
*Don't call the RegisterControllers if its not web
 */
func (c *controllersObject) RegisterControllers() {
	localControllers := configmanager.GetInstance().Controllers
	for i := range localControllers {
		c.registerControllers(localControllers[i], true)
	}
}

func (c *controllersObject) registerControllers(key string, registerApis bool) {
	switch key {
	case baseconst.User:
		c.controllers[key] = &controllers.UserController{BaseControllerFactory: c, ValidatorInterface: &validators.UserValidator{}}
	case baseconst.Session:
		c.controllers[key] = &controllers.SessionController{BaseControllerFactory: c, ValidatorInterface: &validators.SessionValidator{}}
	case baseconst.Roles:
		c.controllers[key] = &controllers.RolesController{BaseControllerFactory: c, ValidatorInterface: &validators.RolesValidator{}}
	case baseconst.Genres:
		c.controllers[key] = &controllers.GenresController{BaseControllerFactory: c, ValidatorInterface: &validators.GenresValidator{}}
	case baseconst.Status:
		c.controllers[key] = &controllers.StatusController{BaseControllerFactory: c, ValidatorInterface: &validators.StatusValidator{}}
	case baseconst.Language:
		c.controllers[key] = &controllers.LanguageController{BaseControllerFactory: c, ValidatorInterface: &validators.LanguageValidator{}}
	case baseconst.TitleType:
		c.controllers[key] = &controllers.TitleTypeController{BaseControllerFactory: c, ValidatorInterface: &validators.TitleTypeValidator{}}
	case baseconst.Titles:
		c.controllers[key] = &controllers.TitlesController{BaseControllerFactory: c, ValidatorInterface: &validators.TitlesValidator{}}
	case baseconst.LanguageMeta:
		c.controllers[key] = &controllers.LanguageMetadataController{BaseControllerFactory: c, ValidatorInterface: &validators.LanguageMetadataValidator{}}
	case baseconst.LanguageContent:
		c.controllers[key] = &controllers.LanguageContentController{BaseControllerFactory: c, ValidatorInterface: &validators.LanguageValidator{}}
	case baseconst.TitlesSummary:
		c.controllers[key] = &controllers.TitlesSummaryController{BaseControllerFactory: c, ValidatorInterface: &validators.TitlesSummaryValidator{}}
	case baseconst.ContentType:
		c.controllers[key] = &controllers.ContentTypeController{BaseControllerFactory: c, ValidatorInterface: &validators.ContentTypeValidator{}}
	case baseconst.Content:
		c.controllers[key] = &controllers.ContentController{BaseControllerFactory: c, ValidatorInterface: &validators.ContentValidator{}}
	case baseconst.TitleMeta:
		c.controllers[key] = &controllers.TitleMetaController{BaseControllerFactory: c, ValidatorInterface: &validators.TitleMetaValidator{}}
	case baseconst.RefreshToken:
		c.controllers[key] = &controllers.RefreshTokensController{BaseControllerFactory: c, ValidatorInterface: &validators.RefreshTokensValidator{}}
	}
	funcs, _ := basefunctions.GetInstance().GetFunctions(basetypes.PSQL, c.controllers[key].GetDBName())
	c.controllers[key].SetBaseFunctions(*funcs)
	c.controllers[key].DoIndexing()
	if registerApis {
		c.controllers[key].RegisterApis()
	}
}
