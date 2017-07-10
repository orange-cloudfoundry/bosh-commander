package main

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
)

func GenerateDirector(boshDirector BoshDirector) (boshdir.Director, error) {
	uaa, err := buildUAA(boshDirector)
	if err != nil {
		return nil, err
	}

	director, err := buildDirector(boshDirector, uaa)
	if err != nil {
		return nil, err
	}
	return director, nil
}

func buildUAA(boshDirector BoshDirector) (boshuaa.UAA, error) {
	if boshDirector.UaaUrl == "" {
		return nil, nil
	}
	factory := boshuaa.NewFactory(loggerBosh)

	// Build a UAA config from a URL.
	// HTTPS is required and certificates are always verified.
	config, err := boshuaa.NewConfigFromURL(boshDirector.UaaUrl)
	if err != nil {
		return nil, err
	}

	// Set client credentials for authentication.
	// Machine level access should typically use a client instead of a particular user.
	config.Client = boshDirector.Username
	config.ClientSecret = boshDirector.Password

	// Configure trusted CA certificates.
	// If nothing is provided default system certificates are used.
	config.CACert = boshDirector.CACert

	return factory.New(config)
}

func buildDirector(boshDirector BoshDirector, uaa boshuaa.UAA) (boshdir.Director, error) {
	factory := boshdir.NewFactory(loggerBosh)

	// Build a Director config from address-like string.
	// HTTPS is required and certificates are always verified.
	config, err := boshdir.NewConfigFromURL(boshDirector.DirectorUrl)
	if err != nil {
		return nil, err
	}

	// Configure custom trusted CA certificates.
	// If nothing is provided default system certificates are used.
	config.CACert = boshDirector.CACert

	// Allow Director to fetch UAA tokens when necessary.
	if uaa != nil {
		config.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc
	} else {
		config.Client = boshDirector.Username
		config.ClientSecret = boshDirector.Password
	}

	return factory.New(config, boshdir.NewNoopTaskReporter(), boshdir.NewNoopFileReporter())
}
