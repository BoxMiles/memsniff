package redis

// dead for now, but categorization of expected result types will be useful for more comprehensive Redis support
/*
type responseParser func(c *Consumer, r *Reader)

var (
	commandMap = map[string]responseParser{
		// Cluster
		"CLUSTER ADDSLOTS":              statusCmd,
		"CLUSTER COUNT-FAILURE-REPORTS": intCmd,
		"CLUSTER COUNTKEYSINSLOT":       intCmd,
		"CLUSTER DELSLOTS":              statusCmd,
		"CLUSTER FAILOVER":              statusCmd,
		"CLUSTER FORGET":                statusCmd,
		"CLUSTER GETKEYSINSLOT":         intCmd,
		"CLUSTER INFO":                  stringCmd,
		"CLUSTER KEYSLOT":               intCmd,
		"CLUSTER MEET":                  statusCmd,
		"CLUSTER NODES":                 stringCmd,
		"CLUSTER REPLICATE":             statusCmd,
		"CLUSTER RESET HARD":            statusCmd,
		"CLUSTER RESET SOFT":            statusCmd,
		"CLUSTER SAVECONFIG":            statusCmd,
		"CLUSTER SET-CONFIG-EPOCH":      statusCmd,
		"CLUSTER SETSLOT":               statusCmd,
		"CLUSTER SLAVES":                stringCmd,
		"CLUSTER SLOTS":                 clusterSlotsCmd,

		// Connection
		"AUTH":   statusCmd,
		"ECHO":   stringCmd,
		"PING":   statusCmd,
		"QUIT":   statusCmd,
		"SELECT": statusCmd,
		"SWAPDB": statusCmd,

		// Geo
		"GEOADD":            intCmd,
		"GEOHASH":           stringSliceCmd,
		"GEOPOS":            geoPosCmd,
		"GEODIST":           floatCmd,
		"GEORADIUS":         geoLocationCmd,
		"GEORADIUSBYMEMBER": geoLocationCmd,

		// Hashes
		"HDEL":         intCmd,
		"HEXISTS":      boolCmd,
		"HGET":         stringCmd,
		"HGETALL":      stringStringMapCmd,
		"HINCRBY":      intCmd,
		"HINCRBYFLOAT": floatCmd,
		"HKEYS":        stringSliceCmd,
		"HLEN":         intCmd,
		"HMGET":        sliceCmd,
		"HMSET":        statusCmd,
		"HSET":         boolCmd,
		"HSETNX":       boolCmd,
		"HSTRLEN":      intCmd,
		"HVALS":        stringSliceCmd,
		"HSCAN":        scanCmd,

		// HyperLogLog
		"PFADD":   intCmd,
		"PFCOUNT": intCmd,
		"PFMERGE": statusCmd,

		// Keys
		"DEL":             intCmd,
		"DUMP":            stringCmd,
		"EXISTS":          intCmd,
		"EXPIRE":          boolCmd,
		"EXPIREAT":        boolCmd,
		"KEYS":            stringSliceCmd,
		"MIGRATE":         statusCmd,
		"MOVE":            boolCmd,
		"OBJECT REFCOUNT": intCmd,
		"OBJECT ENCODING": stringCmd,
		"OBJECT IDLETIME": durationCmd,
		"PERSIST":         boolCmd,
		"PEXPIRE":         boolCmd,
		"PEXPIREAT":       boolCmd,
		"PTTL":            durationCmd,
		"RANDOMKEY":       stringCmd,
		"RENAME":          statusCmd,
		"RENAMENX":        boolCmd,
		"RESTORE":         statusCmd,
		"SORT":            stringSliceCmd,
		"TOUCH":           intCmd,
		"TTL":             durationCmd,
		"TYPE":            statusCmd,
		"UNLINK":          intCmd,
		"WAIT":            intCmd,
		"SCAN":            scanCmd,

		// Lists
		"BLPOP":      stringSliceCmd,
		"BRPOP":      stringSliceCmd,
		"BRPOPLPUSH": stringCmd,
		"LINDEX":     stringCmd,
		"LINSERT":    intCmd,
		"LLEN":       intCmd,
		"LPOP":       stringCmd,
		"LPUSH":      intCmd,
		"LPUSHX":     intCmd,
		"LRANGE":     stringSliceCmd,
		"LREM":       intCmd,
		"LSET":       statusCmd,
		"LTRIM":      statusCmd,
		"RPOP":       stringCmd,
		"RPOPLPUSH":  stringCmd,
		"RPUSH":      intCmd,
		"RPUSHX":     intCmd,

		// Pub/Sub
		//"PSUBSCRIBE": subscribeCmd,
		"PUBSUB CHANNELS": stringSliceCmd,
		"PUBSUB NUMSUB":   stringIntMapCmd,
		"PUBSUB NUMPAT":   intCmd,
		"PUBLISH":         intCmd,
		//"SUBSCRIBE": subscribeCmd,
		// only valid while subscribed
		//"PUNSUBSCRIBE": ,
		//"UNSUBSCRIBE": ,

		// Scripting
		"EVAL":          scriptCmd,
		"EVALSHA":       scriptCmd,
		"SCRIPT DEBUG":  statusCmd,
		"SCRIPT EXISTS": boolSliceCmd,
		"SCRIPT FLUSH":  statusCmd,
		"SCRIPT KILL":   statusCmd,
		"SCRIPT LOAD":   stringCmd,

		// Server
		"BGREWRITEAOF":     statusCmd,
		"BGSAVE":           statusCmd,
		"CLIENT KILL":      statusCmd,
		"CLIENT LIST":      stringCmd,
		"CLIENT GETNAME":   stringCmd,
		"CLIENT PAUSE":     boolCmd,
		"CLIENT REPLY":     statusCmd,
		"CLIENT SETNAME":   statusCmd,
		"COMMAND":          commandsInfoCmd,
		"CONFIG GET":       sliceCmd,
		"CONFIG REWRITE":   statusCmd,
		"CONFIG SET":       statusCmd,
		"CONFIG RESETSTAT": statusCmd,
		"DBSIZE":           intCmd,
		"DEBUG OBJECT":     stringCmd,
		// server will not reply
		//"DEBUG SEGFAULT": ,
		"FLUSHALL":       statusCmd,
		"FLUSHALL ASYNC": statusCmd,
		"FLUSHDB":        statusCmd,
		"FLUSHDB ASYNC":  statusCmd,
		"INFO":           stringCmd,
		"LASTSAVE":       intCmd,
		//"MONITOR": monitorCmd,
		//"ROLE": roleCmd,
		"SAVE":           statusCmd,
		"SHUTDOWN":       statusCmd,
		"SHUTDOWNSAVE":   statusCmd,
		"SHUTDOWNNOSAVE": statusCmd,
		"SLAVEOF":        statusCmd,
		//"SLOWLOG": slowlogCmd,
		// undocumented?
		//"SYNC": ,
		"TIME": timeCmd,

		// Sets
		"SADD":        intCmd,
		"SCARD":       intCmd,
		"SDIFF":       stringSliceCmd,
		"SDIFFSTORE":  intCmd,
		"SINTER":      stringSliceCmd,
		"SINTERSTORE": intCmd,
		"SISMEMBER":   boolCmd,
		"SMEMBERS":    stringSliceCmd,
		"SMOVE":       boolCmd,
		"SPOP":        stringCmd,
		"SRANDMEMBER": stringCmd,
		"SREM":        intCmd,
		"SUNION":      stringSliceCmd,
		"SUNIONSTORE": intCmd,
		"SSCAN":       scanCmd,

		// Sorted Sets
		"ZADD":             intCmd,
		"ZCARD":            intCmd,
		"ZCOUNT":           intCmd,
		"ZINCRBY":          floatCmd,
		"ZINTERSTORE":      intCmd,
		"ZLEXCOUNT":        intCmd,
		"ZRANGE":           stringSliceCmd,
		"ZRANGEBYLEX":      stringSliceCmd,
		"ZREVRANGEBYLEX":   stringSliceCmd,
		"ZRANGEBYSCORE":    stringSliceCmd,
		"ZRANK":            intCmd,
		"ZREM":             intCmd,
		"ZREMRANGEBYLEX":   intCmd,
		"ZREMRANGEBYRANK":  intCmd,
		"ZREMRANGEBYSCORE": intCmd,
		"ZREVRANGE":        stringSliceCmd,
		"ZREVRANGEBYSCORE": stringSliceCmd,
		"ZREVRANK":         intCmd,
		"ZSCORE":           floatCmd,
		"ZUNIONSTORE":      intCmd,
		"ZSCAN":            scanCmd,

		// Strings
		"APPEND":   intCmd,
		"BITCOUNT": intCmd,
		//"BITFIELD": intSliceCmd,
		"BITOP":       intCmd,
		"BITPOS":      intCmd,
		"DECR":        intCmd,
		"DECRBY":      intCmd,
		"GET":         stringCmd,
		"GETBIT":      intCmd,
		"GETRANGE":    stringCmd,
		"GETSET":      stringCmd,
		"INCR":        intCmd,
		"INCRBY":      intCmd,
		"INCRBYFLOAT": floatCmd,
		"MGET":        sliceCmd,
		"MSET":        statusCmd,
		"MSETNX":      boolCmd,
		"PSETEX":      statusCmd,
		"SET":         statusCmd,
		"SETBIT":      intCmd,
		"SETEX":       statusCmd,
		"SETNX":       boolCmd,
		"SETRANGE":    intCmd,
		"STRLEN":      intCmd,
	}
)

func scriptCmd(c *Consumer, r *Reader) {
	r.ReadReply(sliceParser)
}

func sliceCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(sliceParser)
}

func statusCmd(c *Consumer, r *Reader) {
	r.ReadStringReply()
}

func intCmd(c *Consumer, r *Reader) {
	r.ReadIntReply()
}

func durationCmd(c *Consumer, r *Reader) {
	r.ReadIntReply()
}

func timeCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(timeParser)
}

func boolCmd(c *Consumer, r *Reader) {
	r.ReadReply(nil)
}

func stringCmd(c *Consumer, r *Reader) {
	r.ReadBytesReply()
}

func floatCmd(c *Consumer, r *Reader) {
	r.ReadFloatReply()
}

func stringSliceCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(stringSliceParser)
}

func boolSliceCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(boolSliceParser)
}

func stringStringMapCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(stringStringMapParser)
}

func stringIntMapCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(stringIntMapParser)
}

func scanCmd(c *Consumer, r *Reader) {
	r.ReadScanReply()
}

func clusterSlotsCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(clusterSlotsParser)
}

func geoLocationCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(sliceParser)
}

func geoPosCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(geoPosSliceParser)
}

func commandsInfoCmd(c *Consumer, r *Reader) {
	r.ReadArrayReply(commandInfoSliceParser)
}
*/
