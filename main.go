package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

var (
	reServersEndpoint     = regexp.MustCompile("^ZOOKEEPER_SERVER_([0-9]+)_ENDPOINT$")
	reServersLeaderPort   = regexp.MustCompile("^ZOOKEEPER_SERVER_([0-9]+)_LEADER_PORT$")
	reServersElectionPort = regexp.MustCompile("^ZOOKEEPER_SERVER_([0-9]+)_ELECTION_PORT$")
)

type Config struct {
	ClientPort uint16 `env:"ZOOKEEPER_CLIENT_PORT"`
	TickTime   uint16 `env:"ZOOKEEPER_TICK_TIME"`
	InitLimit  uint16 `env:"ZOOKEEPER_INIT_LIMIT"`
	SyncLimit  uint16 `env:"ZOOKEEPER_SYNC_LIMIT"`
	ServerID   uint8  `env:"ZOOKEEPER_SERVER_ID"`
	Servers    []Server
}

type Server struct {
	ServerID     uint8
	Endpoint     string
	LeaderPort   uint16
	ElectionPort uint16
}

func (c *Config) GetEnvClientPort() error {
	if f, ok := reflect.TypeOf(*c).FieldByName("ClientPort"); ok {
		if v, ok := os.LookupEnv(f.Tag.Get("env")); ok {
			i, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return err
			}
			c.ClientPort = uint16(i)
		}
	}
	return nil
}

func (c *Config) GetEnvTickTime() error {
	if f, ok := reflect.TypeOf(*c).FieldByName("TickTime"); ok {
		if v, ok := os.LookupEnv(f.Tag.Get("env")); ok {
			i, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return err
			}
			c.TickTime = uint16(i)
		}
	}
	return nil
}

func (c *Config) GetEnvInitLimit() error {
	if f, ok := reflect.TypeOf(*c).FieldByName("InitLimit"); ok {
		if v, ok := os.LookupEnv(f.Tag.Get("env")); ok {
			i, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return err
			}
			c.InitLimit = uint16(i)
		}
	}
	return nil
}

func (c *Config) GetEnvSyncLimit() error {
	if f, ok := reflect.TypeOf(*c).FieldByName("SyncLimit"); ok {
		if v, ok := os.LookupEnv(f.Tag.Get("env")); ok {
			i, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return err
			}
			c.SyncLimit = uint16(i)
		}
	}
	return nil
}

func (c *Config) GetEnvServerID() error {
	if f, ok := reflect.TypeOf(*c).FieldByName("ServerID"); ok {
		if v, ok := os.LookupEnv(f.Tag.Get("env")); ok {
			i, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return err
			}
			c.ServerID = uint8(i)
		}
	}
	return nil
}

func (c *Config) GetEnvServers() error {
	m := make(map[uint8]*Server)
	for _, k := range os.Environ() {
		v := strings.Split(k, "=")
		if len(v) != 2 {
			return fmt.Errorf("invalid environment variable: %s", v[0])
		}
		if !strings.HasPrefix(v[0], "ZOOKEEPER_SERVER_") {
			continue
		}
		if strings.HasPrefix(v[0], "ZOOKEEPER_SERVER_ID") {
			continue
		}
		if r := reServersEndpoint.FindStringSubmatch(v[0]); len(r) > 0 {
			i, err := strconv.ParseUint(r[1], 10, 8)
			if err != nil {
				return err
			}
			if _, ok := m[uint8(i)]; !ok {
				m[uint8(i)] = &Server{}
			}
			m[uint8(i)].Endpoint = v[1]
		}
		if r := reServersLeaderPort.FindStringSubmatch(v[0]); len(r) > 0 {
			i, err := strconv.ParseUint(r[1], 10, 8)
			if err != nil {
				return err
			}
			if _, ok := m[uint8(i)]; !ok {
				m[uint8(i)] = &Server{}
			}
			n, err := strconv.ParseUint(v[1], 10, 16)
			if err != nil {
				return err
			}
			m[uint8(i)].LeaderPort = uint16(n)
		}
		if r := reServersElectionPort.FindStringSubmatch(v[0]); len(r) > 0 {
			i, err := strconv.ParseUint(r[1], 10, 8)
			if err != nil {
				return err
			}
			if _, ok := m[uint8(i)]; !ok {
				m[uint8(i)] = &Server{}
			}
			n, err := strconv.ParseUint(v[1], 10, 16)
			if err != nil {
				return err
			}
			m[uint8(i)].ElectionPort = uint16(n)
		}
	}
	for k, v := range m {
		if v.Endpoint != "" && v.LeaderPort > 0 && v.ElectionPort > 0 {
			c.Servers = append(c.Servers, Server{
				ServerID:     k,
				Endpoint:     v.Endpoint,
				LeaderPort:   v.LeaderPort,
				ElectionPort: v.ElectionPort,
			})
		}
	}
	sort.Slice(c.Servers, func(i, j int) bool {
		return c.Servers[i].ServerID < c.Servers[j].ServerID
	})
	return nil
}

type output struct{}

func (o *output) Write(b []byte) (int, error) {
	msg := string(b)
	if len(msg) > 0 {
		fmt.Print(msg)
	}
	return len(b), nil
}

func main() {
	cfg := Config{
		ClientPort: 2181,
		TickTime:   2000,
		InitLimit:  10,
		SyncLimit:  5,
	}

	if err := cfg.GetEnvClientPort(); err != nil {
		log.Fatalln(err)
	}

	if err := cfg.GetEnvTickTime(); err != nil {
		log.Fatalln(err)
	}

	if err := cfg.GetEnvInitLimit(); err != nil {
		log.Fatalln(err)
	}

	if err := cfg.GetEnvSyncLimit(); err != nil {
		log.Fatalln(err)
	}

	if err := cfg.GetEnvServerID(); err != nil {
		log.Fatalln(err)
	}

	if err := cfg.GetEnvServers(); err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll("/zookeeper/data", 0755); err != nil {
		log.Fatalln(err)
	}

	if err := syscall.Access("/zookeeper/data", syscall.O_RDWR); err != nil {
		log.Fatalln(err)
	}

	var b strings.Builder

	fmt.Fprintf(&b, "dataDir=/zookeeper/data\n")
	fmt.Fprintf(&b, "clientPort=%v\n", cfg.ClientPort)

	if cfg.TickTime != 0 {
		fmt.Fprintf(&b, "tickTime=%v\n", cfg.TickTime)
	}
	if cfg.InitLimit != 0 {
		fmt.Fprintf(&b, "initLimit=%v\n", cfg.InitLimit)
	}
	if cfg.SyncLimit != 0 {
		fmt.Fprintf(&b, "syncLimit=%v\n", cfg.SyncLimit)
	}

	if cfg.ServerID != 0 {
		for _, s := range cfg.Servers {
			fmt.Fprintf(&b, "server.%v=%v:%v:%v\n", s.ServerID, s.Endpoint, s.LeaderPort, s.ElectionPort)
		}
	}

	if cfg.ServerID != 0 {
		if err := os.WriteFile("/zookeeper/data/myid", []byte(fmt.Sprintf("%v", cfg.ServerID)), 0644); err != nil {
			log.Fatalln(err)
		}
	}

	if err := syscall.Access("/zookeeper/conf/zoo.cfg", syscall.O_RDWR); err == nil {
		if err := os.WriteFile("/zookeeper/conf/zoo.cfg", []byte(b.String()), 0644); err != nil {
			log.Fatalln(err)
		}
	}

	os.Clearenv()
	ctx := context.Background()
	cmd := exec.CommandContext(ctx,
		"/usr/bin/java",
		"-cp", "/zookeeper/lib/*:/zookeeper/conf",
		"org.apache.zookeeper.server.quorum.QuorumPeerMain",
		"/zookeeper/conf/zoo.cfg",
	)
	cmd.Stdout = &output{}
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
