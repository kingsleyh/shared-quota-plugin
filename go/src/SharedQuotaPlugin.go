package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/TykTechnologies/tyk/config"
	logger "github.com/TykTechnologies/tyk/log"
	"github.com/TykTechnologies/tyk/storage"
)

const pluginDefaultKeyPrefix = "apikey-"

// Global redis variables
var conf config.Config
var rc *storage.RedisController

var store = storage.RedisCluster{KeyPrefix: pluginDefaultKeyPrefix}
var log = logger.Get()

func tykGetData(key string) (string, error) {
	val, err := store.GetKey(key)
	return val, err
}

func establishRedisConnection() {
	// Retrieve global configs
	log.Info("SharedQuotaPlugin retrieving global configs")
	conf = config.Global()

	log.Info("SharedQuotaPlugin initializing redis connection")
	// Create a Redis Controller, which will handle the Redis connection for the storage
	rc = storage.NewRedisController(context.Background())

	log.Info("SharedQuotaPlugin setting up redis storage")
	// Create a storage object, which will handle Redis operations using "apikey-" key prefix
	store = storage.RedisCluster{KeyPrefix: pluginDefaultKeyPrefix, HashKeys: conf.HashKeys, RedisController: rc}

	log.Info("SharedQuotaPlugin connecting to redis")
	// Perform Redis connection
	go rc.ConnectToRedis(context.Background(), nil, &conf)
	for i := 0; i < 5; i++ { // max 5 attempts - should only take 2
		time.Sleep(5 * time.Millisecond)
		if rc.Connected() {
			log.Info("Redis Controller connected")
			break
		}
		log.Warn("Redis Controller not connected, will retry")
	}

	// Error handling Redis connection
	if !rc.Connected() {
		log.Error("Could not connect to storage")
		panic("Couldn't esetablished a connection to redis")
	}
}

func init() {
	log.Info("SharedQuotaPlugin init trying to get redis connection")
	establishRedisConnection()
	log.Info("--- SharedQuotaPlugin init success! ---- ")
}

// Get key from the header and retrieve the quota key from redis and replace it in the header
func ApplySharedKey(rw http.ResponseWriter, r *http.Request) {
	log.Info("Start SharedQuota Plugin")

	var redisKey = r.Header.Get("Authorization")
	log.Info("SharedQuota: ", redisKey, " retrieved from header")

	val, err := tykGetData(redisKey)
	if err != nil {
		log.Error("Plugin redis error ", err)
	}
	log.Info("SharedQuota: value retrieved")

	if !json.Valid([]byte(val)) {
		log.Error("SharedQuota: invalid json")
		return
	}

	var result map[string]any
	json.Unmarshal([]byte(val), &result)

	// get the new key from the result in meta_data field quota_key
	_, ok := result["meta_data"]
	if !ok {
		log.Error("SharedQuota: meta_data not found")
		return
	}
	meta_data := result["meta_data"].(map[string]any)

	_, ok = meta_data["quota_token"]
	if !ok {
		log.Error("SharedQuota: quota_token not found")
		return
	}

	newKey := meta_data["quota_token"].(string)
	log.Info("SharedQuota: new key retrieved: " + newKey)

	// set the new key in the header
	r.Header.Set("Authorization", newKey)
	log.Info("SharedQuota: new key set in header: " + newKey)

	log.Info("End SharedQuota Plugin")

}

func main() {
}
