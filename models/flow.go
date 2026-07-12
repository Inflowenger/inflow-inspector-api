package models

import (
	inflowModels "github.com/Inflowenger/inflow-fusion/compilers/vueFlow"
)

type FlowRecord struct {
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	CreatedAt int64                `json:"createdAt"`
	UpdatedAt int64                `json:"updatedAt"`
	ViewFlow  inflowModels.VueFlow `json:"view_flow"`
}
