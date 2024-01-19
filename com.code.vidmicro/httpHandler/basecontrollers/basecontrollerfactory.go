package basecontrollers

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/app/controllers"
	"com.code.vidmicro/com.code.vidmicro/app/validators"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
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
	case User:
		c.controllers[key] = &controllers.UserController{BaseControllerFactory: c, ValidatorInterface: &validators.UserValidator{}}
	case Session:
		c.controllers[key] = &controllers.SessionController{BaseControllerFactory: c, ValidatorInterface: &validators.SessionValidator{}}
	}
	funcs, _ := basefunctions.GetInstance().GetFunctions(basetypes.MYSQL, c.controllers[key].GetDBName())
	c.controllers[key].SetBaseFunctions(*funcs)
	c.controllers[key].DoIndexing()
	if registerApis {
		c.controllers[key].RegisterApis()
	}
}
