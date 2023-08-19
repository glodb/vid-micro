package basevalidators

type ValidatorInterface interface {
	Validate(apiName string, data interface{}) error
}
