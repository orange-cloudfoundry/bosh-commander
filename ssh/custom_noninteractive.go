package ssh

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type CustomNonInteractiveRunner struct {
	customRunner CustomRunner
}

func NewCustomNonInteractiveRunner(customRunner CustomRunner) CustomNonInteractiveRunner {
	return CustomNonInteractiveRunner{customRunner}
}

func (r CustomNonInteractiveRunner) Run(connOpts boshssh.ConnectionOpts, result boshdir.SSHResult, rawCmd []string) error {
	if len(result.Hosts) == 0 {
		return bosherr.Errorf("Non-interactive SSH expects at least one host")
	}
	if len(rawCmd) == 0 {
		return bosherr.Errorf("Non-interactive SSH expects non-empty command")
	}
	cmdFactory := func(host boshdir.Host) boshsys.Command {
		return boshsys.Command{
			Name:         "ssh",
			KeepAttached: true,
			Args:         append([]string{host.Host, "-l", host.Username}, rawCmd...),
		}
	}

	return r.customRunner.Run(connOpts, result, cmdFactory)
}
