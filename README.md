Allows you to load the system env into structures automatically  


example:

To use you must import the following:

```go
import "github.com/gojetpack/goconf"
```

Simple example:

In the following example you can see that there are two configuration structures:
- RedisConfig
- MongoConfig

All Redis environment variables will have the REDIS_ prefix in front. i.e: REDIS_HOST

All Mongo environment variables will have the MONGO_ prefix in front. i.e: MONGO_HOST 

You can associate an environment variable to a field as follows:

- Adding the "env" tag to the field, followed by the name of the environment variable. i.e:
```go    
    SecureConnection bool `env:"SECURE_CONNECTION"`
 ```
- By the field name. By default the field name is converted to "Screaming Snake Case" to match with the env var. 
i.e: SecureConnection -> SECURE_CONNECTION

The configuration structure is "ExtractorArgs"

```go
type ExtractorArgs struct {
	// Allow to configure the env extraction
	Options ExtractorOptions

	// It must be an even array of elements.
	// For each tuple:
	//  - The first element will be a pointer to the object in which the configuration will be saved.
	//  - The second element will be the prefix for this configuration
	Configs []interface{}
}
```

```go
type RedisConfig struct {
    Host             string
    Port             int
    AuthToken        string
    SecureConnection bool `env:"SECURE_CONNECTION"`
}
redisConfig := RedisConfig{}

type MongoConfig struct {
    Host             string
    Port             int
    DatabaseName     string
    Test             bool
    SecureConnection bool `env:"SECURE_CONNECTION"`
}
mongoConfig := MongoConfig{}


args := goconf.ExtractorArgs{
    Options: goconf.ExtractorOptions{
        EnvFile: "testdata/env_with_prefix", // env file path
    },
    Configs: []interface{}{
    //  Config struct | env name prefix
        &redisConfig, "REDIS",
        &mongoConfig, "MONGO",
    },
}
err := goconf.Extract(args)
```

Mapping

![](doc/map.png)

Prefix association

![](doc/prefix.png)

Result

![](doc/result.png)