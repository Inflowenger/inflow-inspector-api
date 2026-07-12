package repository

import (
	"fmt"

	"github.com/Inflowenger/inflow-inspector-api/models"
)

func ContextIndexByInt(i uint64) string {
	return fmt.Sprintf("%s:%d", CONTEXT_INDEX_PREFIX, i)
}
func ContextIndexByString(i string) string {
	return fmt.Sprintf("%s:%s", CONTEXT_INDEX_PREFIX, i)
}
func UpsertContext(f *models.ContextRecord) error {
	if f.ID == "" {
		index, err := Seq()
		if err != nil {
			return err
		}
		f.ID = ContextIndexByInt(index)
	} else {
		db := GetBadgerDb(models.ContextRecord{})
		rec, err := db.Get(f.ID)
		if err != nil {
			return err
		}
		f.CreatedAt = rec.CreatedAt

	}

	return UpsertWithKeys([]string{f.ID}, f)
}

func GetContextById(key string) *models.ContextRecord {
	db := GetBadgerDb(models.ContextRecord{})
	f, _ := db.Get(key)
	return f
}

func GetContextList(last string, limit int) ([]models.ContextRecord, string, error) {
	db := GetBadgerDb(models.ContextRecord{})

	return db.ListValues(CONTEXT_INDEX_PREFIX, last, limit, true)

}
