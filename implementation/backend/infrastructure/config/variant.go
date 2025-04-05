package config

type AppVariant string

const (
	AppVariant__Radar AppVariant = "radar"
	AppVariant__PGP   AppVariant = "pgp"
	AppVariant__EDA   AppVariant = "eda"
)

// todo test variants
// Database: Postgres, CitusData, YugabyteDB (3)
// Flow control: Early dropper + async
// Test cases
// Postgres + No Flow Control
// CitusData + No Flow Control
// YugabyteDB + No Flow Control
// CitusData + With Flow Control
// YugabyteDB + With Flow Control
