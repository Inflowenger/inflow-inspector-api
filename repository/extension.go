package repository

import (
	"fmt"

	"github.com/Inflowenger/inflow-inspector-api/models"
)

func ExtensionIndexByInt(i uint64) string {
	return fmt.Sprintf("%s:%d", EXTENSION_INDEX_PREFIX, i)
}
func ExtensionIndexByString(i string) string {
	return fmt.Sprintf("%s:%s", EXTENSION_INDEX_PREFIX, i)
}
func UpsertExtension(extInstance *models.ExtensionRecord) error {
	if extInstance.ID == "" {
		index, err := Seq()
		if err != nil {
			return err
		}
		extInstance.ID = ExtensionIndexByInt(index)
	} else {
		db := GetBadgerDb(models.ExtensionRecord{})
		rec, err := db.Get(extInstance.ID)
		if err != nil {
			return err
		}
		extInstance.CreatedAt = rec.CreatedAt

	}

	return UpsertWithKeys([]string{extInstance.ID}, extInstance)
}

func GetExtensionById(key string) (*models.ExtensionRecord, error) {
	db := GetBadgerDb(models.ExtensionRecord{})
	return db.Get(key)

}

func GetExtensionList(last string, limit int) ([]models.ExtensionRecord, string, error) {
	db := GetBadgerDb(models.ExtensionRecord{})

	return db.ListValues(EXTENSION_INDEX_PREFIX, last, limit, true)

}
