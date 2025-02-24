package postgres

type Cluster struct {
	Leader    *Base
	Followers []Base
}

// todo update the config and initialization function
