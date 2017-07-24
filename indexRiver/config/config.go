// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Mysqlbeat MysqlbeatConfig
}

type MysqlbeatConfig struct {
	Period            string   `yaml:"period"`
	Hostname          string   `yaml:"hostname"`
	Port              string   `yaml:"port"`
	Username          string   `yaml:"username"`
	Password          string   `yaml:"password"`
	EncryptedPassword string   `yaml:"encryptedpassword"`
	Queries           []string `yaml:"queries"`
	QueryTypes        []string `yaml:"querytypes"`
	DeltaWildcard     string   `yaml:"deltawildcard"`
	DeltaKeyWildcard  string   `yaml:"deltakeywildcard"`
	RedisServer 	  string   `yaml:"redisserver"`
	RedisUsername 	  string   `yaml:"redisusername"`
	RedisPassword 	  string   `yaml:"redispassword"`
	RedisRiverkey 	  string   `yaml:"redisriverkey"`
	EsServer 	  	  string   `yaml:"esserver"`
	EsUsername 	  	  string   `yaml:"esusername"`
	EsPassword 	  	  string   `yaml:"espassword"`
	EsSniff 	  	  bool     `yaml:"essniff"`
	EsBulkSize 	  	  int     `yaml:"esbulksize"`
	Suppliers 	  	  []string `yaml:"suppliers"`
	SupQueries 	  	  []string `yaml:"supqueries"`	
}
