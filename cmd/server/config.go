package server

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Addr      string
	DataDir   string
	SelfCheck bool
}

func ParseConfig(args []string) Config {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:19081", "监听地址")
	data := fs.String("data", ".refill-data", "数据目录")
	self := fs.Bool("selfcheck", false, "运行有界自检")
	_ = fs.Parse(args)
	if env := os.Getenv("PORT"); env != "" {
		*addr = "127.0.0.1:" + env
	}
	return Config{Addr: *addr, DataDir: *data, SelfCheck: *self}
}
func ValidateConfig(c Config) error {
	if c.Addr == "" {
		return fmt.Errorf("监听地址不能为空")
	}
	return ValidatePort(c.Addr)
}
