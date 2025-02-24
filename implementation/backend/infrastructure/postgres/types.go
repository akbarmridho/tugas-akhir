package postgres

type Config struct {
	DatabaseUsername string
	DatabasePassword string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	Timezone         *string
}
