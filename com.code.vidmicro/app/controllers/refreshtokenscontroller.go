package controllers

import (
	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/chenzhuoyu/base64x"
)

type RefreshTokensController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
	encoding base64x.Encoding
}

func (u RefreshTokensController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u RefreshTokensController) GetCollectionName() basetypes.CollectionName {
	return "refresh_tokens"
}

func (u RefreshTokensController) DoIndexing() error {
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.RefreshToken{})
	return nil
}

func (u *RefreshTokensController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *RefreshTokensController) GetRefreshToken(userId int) (string, error) {
	stringToken := ""
	rowsCount := 0

	selectRows, err := u.Find(u.GetDBName(), u.GetCollectionName(), " refresh_token", map[string]interface{}{"user_id": userId}, &models.RefreshToken{}, true, "", false)

	if err != nil {
		return "", err
	}

	defer selectRows.Close()
	for selectRows.Next() {
		err = selectRows.Scan(&stringToken)
		if err != nil {
			return "", err
		}
		rowsCount += 1
	}

	if rowsCount > 0 {
		return stringToken, nil
	}

	stringToken, _ = utils.GenerateUUID()
	stringToken = u.encoding.EncodeToString([]byte(stringToken))
	token := models.RefreshToken{
		RefreshToken: stringToken,
		UserId:       userId,
	}

	_, err = u.Add(u.GetDBName(), u.GetCollectionName(), token, false)
	if err != nil {
		return "", err
	}
	return stringToken, nil
}

func (u *RefreshTokensController) ValidateRefreshToken(token string) (bool, error) {

	rowsCount, err := u.Count(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"refresh_token": token})

	if err != nil {
		return false, err
	}

	if rowsCount > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (u RefreshTokensController) RegisterApis() {
}
