package config

type HTTPConfig struct {
    Port string `env:"HTTP_PORT" default:"8080"`
}

type DBConfig struct {
    Host     string `env:"DB_HOST"`
    Port     string `env:"DB_PORT"`
    User     string `env:"DB_USER"`
    Password string `env:"DB_PASSWORD"`
    Name     string `env:"DB_NAME"`
    SSLMode  string `env:"DB_SSLMODE" default:"disable"`
}

type JWTConfig struct {
    Secret    string `env:"JWT_SECRET"`
    ExpireSec int    `env:"JWT_EXPIRE_SEC" default:"3600"`
}

type Config struct {
    Env   string    `env:"ENV" default:"dev"`
    HTTP  HTTPConfig
    DB    DBConfig
    JWT   JWTConfig
    LogLevel string `env:"LOG_LEVEL" default:"debug"`
}