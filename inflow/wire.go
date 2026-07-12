package inflow

import (
	"fmt"
	"strings"

	inflowModels "github.com/Inflowenger/inflow-fusion/models"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/Inflowenger/inflow-inspector-api/repository"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
)

type InflowWire struct{}

func (isvc *InflowWire) RetrieveContext(msg *nats.Msg) {
	fmt.Println(msg.Header)
	// get context data from db
	parts := strings.Split(msg.Subject, ".")
	ctxId := parts[len(parts)-1]
	rec := repository.GetContextById(ctxId)
	if rec == nil {
		msg.Respond([]byte(`{}`))
		return
	}
	wantedContext := inflowModels.ContextDoc{Header: rec.Header, Data: rec.Context}
	b, _ := sonic.Marshal(wantedContext)
	msg.Respond(b)
	// msg.Respond([]byte(`{"header":{},"data":"{\"node1\":{\"b\":2,\"sum\":3,\"a\":1},\"node2\":{\"a\":1,\"b\":2,\"sum\":3}}"}`))
}

func (isvc *InflowWire) UpdateContext(msg *nats.Msg) {
	fmt.Println(string(msg.Data)) // save to db
	contextId := msg.Header.Get("contextId")
	fmt.Println("recieve Update on : ", contextId)
	newCtxData := inflowModels.ContextDoc{}
	err := sonic.Unmarshal(msg.Data, &newCtxData)
	if err != nil {
		fmt.Printf("error in recieved context data with error : %s\n data : %s ", err.Error(), string(msg.Data))
		return
	}
	if strings.HasPrefix(contextId, repository.CONTEXT_INDEX_PREFIX) {
		ctxRecord := repository.GetContextById(contextId)
		if ctxRecord == nil {
			fmt.Printf("given context id not found \ndata : %s ", string(msg.Data))
			return
		}
		ctxRecord.Context = newCtxData.Data
		ctxRecord.Header = newCtxData.Header
		ctxRecord.UpdatedBy = models.LastChange{By: models.ByFlow, Address: msg.Header.Get("flowId")}

		repository.UpsertContext(ctxRecord)

	}
	msg.Respond([]byte(`accepted`))
}

func (isvc *InflowWire) RetrieveFlow(msg *nats.Msg) {
	//pattern of get_flow is inflow.req.flow.get.{flowId}
	// get flowId
	parts := strings.Split(msg.Subject, ".")
	flowId := parts[len(parts)-1]
	rec, err := repository.GetFlowById(flowId)
	if err != nil {
		msg.Respond([]byte(`{}`))
		fmt.Println("given flow id not found or exception error occurred")
	}
	_, cmp, err := FLowCompiler(*rec)
	wantedFlow := inflowModels.Flow{
		UUID:  flowId,
		Nodes: []inflowModels.Node{},
	}
	for _, el := range cmp {
		wantedFlow.Nodes = append(wantedFlow.Nodes, *el)
	}
	b, _ := sonic.Marshal(wantedFlow)

	msg.Respond(b)

}
