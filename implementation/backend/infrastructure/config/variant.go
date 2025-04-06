package config

type DBVariant string

const (
	DBVariant__Postgres   DBVariant = "postgres"
	DBVariant__Citusdata  DBVariant = "citusdata"
	DBVariant__YugabyteDB DBVariant = "yugabytedb"
)

type FlowControlVariant string

const (
	FlowControlVariant__NoFlowControl FlowControlVariant = "no-flow-control"
	FlowControlVariant__DropperAsync  FlowControlVariant = "dropper-async"
)
