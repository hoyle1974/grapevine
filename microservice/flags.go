package microservice

import "flag"

var (
	port = flag.Int("port", 8080, "The server port")
	ip   = flag.String("ip", "0.0.0.0", "address")
	env  = flag.String("env", "dev", "environment")

	dbhost     = flag.String("dbhost", "localhost", "db hostname")
	dbport     = flag.Int("dbport", 5432, "The db port")
	dbuser     = flag.String("dbuser", "postgres", "db user")
	dbpassword = flag.String("dbpassword", "GAVxFWGABz", "db password")
	dbname     = flag.String("dbname", "grapevine", "db name")
)
