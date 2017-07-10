package main

import (
	"bufio"
	"bytes"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/mitchellh/colorstring"
	"github.com/orange-cloudfoundry/bosh-commander/ssh"
	"io/ioutil"
	"strconv"
	"strings"
)

type CommandRunner struct {
	boshCommanderScript *BoshCommanderScript
	uuidGen             boshuuid.Generator
	provider            ssh.CustomProvider
	bufferedResult      *bytes.Buffer
}

func NewCommandRunner(boshCommanderScript *BoshCommanderScript) *CommandRunner {
	bufferedResult := new(bytes.Buffer)
	provider := ssh.NewCustomProvider(boshsys.NewExecCmdRunner(loggerBosh), boshsys.NewOsFileSystem(loggerBosh), bufferedResult, ioutil.Discard, loggerBosh)
	agent := &CommandRunner{
		bufferedResult:      bufferedResult,
		provider:            provider,
		boshCommanderScript: boshCommanderScript,
	}
	return agent
}

func (a *CommandRunner) Run(director BoshDirector) error {
	boshDirector, err := GenerateDirector(director)
	if err != nil {
		return err
	}
	gatewayDisabled := true
	gateway := director.Gateway
	if gateway.Host != "" {
		gatewayDisabled = false
	}
	connOpts := boshssh.ConnectionOpts{
		GatewayDisable:        gatewayDisabled,
		GatewayUsername:       gateway.Username,
		GatewayHost:           gateway.Host,
		GatewayPrivateKeyPath: gateway.PrivateKeyPath,
		RawOpts:               []string{},
	}
	boshSshInstances, err := a.FindBoshSshInstances(
		boshDirector,
		a.boshCommanderScript.JobMatch,
		a.boshCommanderScript.Deployments...,
	)
	if err != nil {
		return err
	}
	scripts := a.privilegeIfNeeded(a.boshCommanderScript.Script)
	afterAll := a.privilegeIfNeeded(a.boshCommanderScript.AfterAll)

	err = a.RunCommandByInstances(boshSshInstances, scripts, connOpts)
	if err != nil {
		return err
	}
	if len(afterAll) == 0 {
		return nil
	}
	err = a.RunCommandByInstances(boshSshInstances, afterAll, connOpts)
	if err != nil {
		return err
	}
	return nil
}
func (a CommandRunner) privilegeIfNeeded(commands []string) []string {
	if !a.boshCommanderScript.Sudo || len(commands) == 0 {
		return commands
	}
	for i, command := range commands {
		commands[i] = "sudo " + command
	}
	return commands
}
func (a *CommandRunner) RunCommandByInstances(boshSshInstances []BoshSshInstance, commands []string, connOpts boshssh.ConnectionOpts) error {
	for _, boshSshInstance := range boshSshInstances {
		err := a.RunCommandByInstance(boshSshInstance, commands, connOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
func (a *CommandRunner) RunCommandByInstance(boshSshInstance BoshSshInstance, commands []string, connOpts boshssh.ConnectionOpts) error {
	instance := boshSshInstance.Instance
	depl := boshSshInstance.Deployment
	if !instance.IsRunning() {
		return nil
	}
	uuidBosh := boshuuid.NewGenerator()
	sshOpts, privKey, err := boshdir.NewSSHOpts(uuidBosh)
	if err != nil {
		return err
	}
	indexJob := strconv.Itoa(*instance.Index)
	slug := boshdir.NewAllOrInstanceGroupOrInstanceSlug(instance.JobName, indexJob)
	sshResult, err := depl.SetUpSSH(slug, sshOpts)
	defer func() {
		_ = depl.CleanUpSSH(slug, sshOpts)
		a.bufferedResult.Reset()
	}()
	if err != nil {

		return err
	}
	connOpts.PrivateKey = privKey
	runner := a.provider.NewSSHRunner()
	for _, command := range commands {
		logger.Info(colorstring.Color("Running command '[green]%s[cyan]' in %s"), command, boshSshInstance.String())
		err = runner.Run(connOpts, sshResult, a.createCommandSlice(command))
		if err != nil {
			if cerr, ok := err.(ssh.ErrCommandSsh); ok {
				a.SshCommandErrToLog(cerr)
				return nil
			}
			return err
		}
		outputSsh := a.bufferedResult.String()
		a.SshOutputToLog(outputSsh)
		a.bufferedResult.Reset()
	}

	return nil
}
func (a *CommandRunner) createCommandSlice(command string) []string {
	splitCommands := strings.Split(command, " ")
	finalCommands := make([]string, 0)
	text := ""
	identifier := ""
	for _, splitCommand := range splitCommands {
		splitCommand = strings.TrimSpace(splitCommand)
		if !strings.HasPrefix(splitCommand, "\"") && !strings.HasPrefix(splitCommand, "'") && text == "" {
			finalCommands = append(finalCommands, splitCommand)
			continue
		}
		if identifier == "" {
			identifier = string(splitCommand[0])
			text = splitCommand
		} else {
			text += " " + splitCommand
		}
		if len(splitCommand) > 1 && string(splitCommand[len(splitCommand)-1]) == identifier {
			identifier = ""
			finalCommands = append(finalCommands, text)
			text = ""
			continue
		}
	}
	return finalCommands
}

func (a *CommandRunner) FindBoshSshInstances(boshDirector boshdir.Director, jobName Regexp, inDeplNames ...Regexp) ([]BoshSshInstance, error) {
	boshSshInstances := make([]BoshSshInstance, 0)
	var deployments []boshdir.Deployment
	var err error
	deployments, err = boshDirector.Deployments()

	if err != nil {
		return boshSshInstances, err
	}
	for _, deployment := range deployments {
		if len(inDeplNames) > 0 && !Regexps(inDeplNames).MatchString(deployment.Name()) {
			continue
		}
		vms, err := deployment.VMInfos()
		if err != nil {
			return boshSshInstances, err
		}
		for _, vm := range vms {
			indexJob := strconv.Itoa(*vm.Index)
			slug := vm.JobName + "/" + indexJob
			if !jobName.MatchString(slug) {
				continue
			}
			boshSshInstances = append(boshSshInstances, BoshSshInstance{
				Deployment: deployment,
				Instance:   vm,
			})
		}
	}
	return boshSshInstances, nil
}

func (a *CommandRunner) SshCommandErrToLog(err ssh.ErrCommandSsh) {
	logOutput := "Error on command :[red]\n"
	logOutput += err.String() + "\n"
	logger.Warning(colorstring.Color(logOutput))
}
func (a *CommandRunner) SshOutputToLog(output string) {
	logOutput := "Command result :[reset]\n"
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		logOutput += line + "\n"
	}
	logger.Info(colorstring.Color(logOutput))
}
