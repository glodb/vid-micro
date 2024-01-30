package validators

type TitlesValidator struct {
}

func (u *TitlesValidator) GetRules(apiName string) map[string]interface{} {
	switch apiName {
	case "/api/titles/put":
	case "/api/titles/delete":
		fallthrough
	case "/api/titles/post":
	}
	return nil
}
