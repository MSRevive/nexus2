package config

import "flag"

type Flags struct {
	Address       string
	Port          int
	ConfigFile    string
	MigrateConfig bool
	Debug         bool
}

func InitializeFlags(args []string) Flags {
	var f Flags

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.StringVar(&f.Address, "addr", "127.0.0.1", "The address of the server.")
	flagSet.IntVar(&f.Port, "port", 1337, "The port this should run on.")
	flagSet.StringVar(&f.ConfigFile, "cfile", "./runtime/config.yaml", "Location of via config file")
	flagSet.BoolVar(&f.Debug, "d", false, "Run with debug mode.")
	flagSet.BoolVar(&f.MigrateConfig, "m", false, "Migrate the ini/toml config to YAML")
	flagSet.Parse(args[1:])

	return f
}
