package validators

import (
	"log"

	"com.code.vidmicro/com.code.vidmicro/app/models"
)

type TitlesSummaryValidator struct {
}

func (u *TitlesSummaryValidator) Validate(apiName string, data interface{}) error {
	titleSummaryData := data.(models.TitlesSummary)
	log.Println(titleSummaryData)
	switch apiName {
	}
	return nil
}
