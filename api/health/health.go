package health

type PingResponse struct {
	App           string `json:"app"`
	Version       string `json:"version"`
	DatabaseReady bool   `json:"database_ready"`
	RedisReady    bool   `json:"redis_ready"`
	Timestamp     int64  `json:"timestamp"`
}
