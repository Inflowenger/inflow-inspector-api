package repository

import (
	"fmt"

	"github.com/Inflowenger/inflow-inspector-api/models"
)

func FlowIndexByInt(i uint64) string {
	return fmt.Sprintf("%s:%d", FLOW_INDEX_PREFIX, i)
}
func FlowIndexByString(i string) string {
	return fmt.Sprintf("%s:%s", FLOW_INDEX_PREFIX, i)
}
func UpsertFlow(f *models.FlowRecord) error {
	if f.ID == "" {
		index, err := Seq()
		if err != nil {
			return err
		}
		f.ID = FlowIndexByInt(index)
	} else {
		db := GetBadgerDb(models.FlowRecord{})
		rec, err := db.Get(f.ID)
		if err != nil {
			return err
		}
		f.CreatedAt = rec.CreatedAt

	}

	return UpsertWithKeys([]string{f.ID}, f)
}

func GetFlowById(key string) (*models.FlowRecord, error) {
	db := GetBadgerDb(models.FlowRecord{})
	return db.Get(key)

}

func GetFlowList(last string, limit int) ([]models.FlowRecord, string, error) {
	db := GetBadgerDb(models.FlowRecord{})

	return db.ListValues(FLOW_INDEX_PREFIX, last, limit, true)

}
