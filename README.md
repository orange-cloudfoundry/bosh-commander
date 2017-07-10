# Bosh-commander

Run a set of commands on multiple vms found by deployments, jobs name and bosh directors.

## Why ?

This was primary created to be able to fix `consul` and `etcd` clusters on a cloud foundry 
deployment when cluster wasn't shutdown gracefully.

## Installation

### On *nix system

You can install this via the command-line with either `curl` or `wget`.

#### via curl

```bash
$ sh -c "$(curl -fsSL https://raw.github.com/orange-cloudfoundry/bosh-commander/master/bin/install.sh)"
```

#### via wget

```bash
$ sh -c "$(wget https://raw.github.com/orange-cloudfoundry/bosh-commander/master/bin/install.sh -O -)"
```

### On windows

You can install it by downloading the `.exe` corresponding to your cpu from releases page: https://github.com/orange-cloudfoundry/bosh-commander/releases .
Alternatively, if you have terminal interpreting shell you can also use command line script above, it will download file in your current working dir.

### From go command line

Simply run in terminal:

```bash
$ go get github.com/orange-cloudfoundry/bosh-commander
```

## Usage

You will need to set available directors by creating a `.bosh_commander.yml` config file in your home directory:

```yml
log_level: INFO # Can be also ERROR, WARN or DEBUG
bosh_directors:
- name: mybosh
  director_url: https://127.0.0.1:25555
  username: myboshusername
  password: myboshpassword
  uaa_url: https://127.0.0.1:8443 # if you use use an uaa url, it will use uaa to authenticate user
  ca_cert_file: path/to/pem.pem
```

You will now have to create a script file to perform commands on your vm following this schema:

```yml
job_match: myjob # this is the name of your job, it can be a regex
sudo: true # to run your commands in privilegied mode
deployments: # This is optionnal, if set it will looking only deployments which match regex given
- mydeployment.*
script: # It will be the set of command you want to run on your vms, if command fail it will continue to perform the next command
- echo "my super command"
after_all: # this is optionnal, this is a set of commands to run after all commands in script have been ran in all vms
- echo "this is ran after all vms ran scripts commands"
```

You can found useful scripts in [/scripts](/scripts).

You can now run `bosh-commander run -f myscript.yml`.