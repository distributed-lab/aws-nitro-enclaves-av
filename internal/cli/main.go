package cli

import (
	"github.com/alecthomas/kingpin"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/config"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/service"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
)

func Run(args []string) bool {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	cfg := config.New(kv.MustFromEnv())
	log = cfg.Log()

	app := kingpin.New("aws-nitro-enclaves-av", "")

	runCmd := app.Command("run", "run command")
	serviceCmd := runCmd.Command("service", "run service") // you can insert custom help

	// custom commands go here...

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return false
	}

	switch cmd {
	case serviceCmd.FullCommand():
		service.Run(cfg)
	// handle any custom commands here in the same way
	default:
		log.Errorf("unknown command %s", cmd)
		return false
	}

	return true
}
