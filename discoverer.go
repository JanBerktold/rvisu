package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-redis/redis"
)

type RedisRole int32

const (
	Master RedisRole = iota
	Slave
	Sentinel
)

type RedisInstance struct {
	Address string
	Role    RedisRole

	Master          *RedisInstance
	SentinelMasters []*RedisInstance
	Slaves          []*RedisInstance
}

type Discoverer struct {
	nodes            map[string]*RedisInstance
	initialAddresses []string
	password         string
}

func NewDiscoverer(initialAddresses []string, password string) *Discoverer {
	return &Discoverer{
		nodes:            map[string]*RedisInstance{},
		initialAddresses: initialAddresses,
		password:         password,
	}
}

func (d *Discoverer) BuildGraph() error {
	for _, address := range d.initialAddresses {
		d.discoverAddress(address)
	}

	return nil
}

func (d *Discoverer) Result() []*RedisInstance {
	result := make([]*RedisInstance, len(d.nodes))
	index := 0

	for _, instance := range d.nodes {
		result[index] = instance
		index++
	}

	return result
}

func (d *Discoverer) discoverAddress(address string) *RedisInstance {
	if instance, ok := d.nodes[address]; ok {
		return instance
	}

	opts := &redis.Options{
		Addr:     address,
		Password: d.password,
	}

	client := redis.NewClient(opts)

	result := client.Info()
	if err := result.Err(); err != nil {
		fmt.Println(err)
		return nil
	}

	info := parseRedisInfo(result.String())
	instance := &RedisInstance{
		Address: address,
		Role:    parseRole(info["role"]),
	}

	d.nodes[address] = instance

	if masterHost, ok := info["master_host"]; ok {
		masterPort := info["master_port"]
		instance.Master = d.discoverAddress(fmt.Sprintf("%v:%v", masterHost, masterPort))
	}

	if numberSlaves, ok := info["connected_slaves"]; ok {
		num, _ := strconv.ParseInt(numberSlaves, 10, 64)

		instance.Slaves = make([]*RedisInstance, num)
		for i := int64(0); i < num; i++ {
			slaveConfig := info["slave"+strconv.FormatInt(i, 10)]
			instance.Slaves[i] = d.discoverAddress(parseSlaveConfig(slaveConfig))
		}
	}

	if numberSentinelMasters, ok := info["sentinel_masters"]; ok {
		num, _ := strconv.ParseInt(numberSentinelMasters, 10, 64)

		instance.SentinelMasters = make([]*RedisInstance, num)
		for i := int64(0); i < num; i++ {
			sentinelConfig := info["master"+strconv.FormatInt(i, 10)]
			instance.SentinelMasters[i] = d.discoverAddress(parseSentinelConfig(sentinelConfig))
		}
	}
	return instance
}

func parseSlaveConfig(config string) string {
	parts := strings.Split(config, ",")
	ip := parts[0][3:]
	port := parts[1][5:]
	return fmt.Sprintf("%v:%v", ip, port)
}

func parseSentinelConfig(config string) string {
	parts := strings.Split(config, ",")
	return parts[2][8:]
}

func parseRole(role string) RedisRole {
	switch role {
	case "master":
		return Master
	case "slave":
		return Slave
	default:
		return Sentinel
	}
}
func parseRedisInfo(in string) map[string]string {
	out := make(map[string]string)
	lines := strings.Split(in, "\r\n")
	for _, line := range lines {
		trimmed := strings.TrimFunc(line, unicode.IsSpace)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) < 2 {
			continue
		}

		out[parts[0]] = parts[1]
	}
	return out
}
