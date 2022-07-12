package configs

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
	"runtime"
)

/**
This is an initial implementation of Config. Please keep in mind we're still working to improve how we handle
environment variables and general config data.
*/

// Config host application configuration
// todo: some of these values can be moved to the .env
type Config struct {
	AwsProfile        string `env:"AWS_PROFILE"`
	LocalOs           string
	LocalArchitecture string
	InstallerEmail    string

	KubefirstLogPath  string `env:"KUBEFIRST_LOG_PATH" envDefault:"logs"`
	HomePath          string
	KubectlClientPath string
	KubeConfigPath    string
	HelmClientPath    string
	TerraformPath     string

	KubectlVersion   string `env:"KUBECTL_VERSION" envDefault:"v1.20.0"`
	TerraformVersion string
	HelmVersion      string

	// todo: move it back
	KubefirstVersion string
}

func ReadConfig() *Config {
	config := Config{}

	if err := env.Parse(&config); err != nil {
		log.Println("something went wrong loading the environment variables")
		log.Panic(err)
	}

	var err error
	config.HomePath, err = os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	config.LocalOs = runtime.GOOS
	config.LocalArchitecture = runtime.GOARCH

	config.KubectlClientPath = fmt.Sprintf("%s/.kubefirst/tools/kubectl", config.HomePath)
	config.KubeConfigPath = fmt.Sprintf("%s/.kubefirst/gitops/terraform/base/kubeconfig_kubefirst", config.HomePath)
	config.TerraformPath = fmt.Sprintf("%s/.kubefirst/tools/terraform", config.HomePath)
	config.HelmClientPath = fmt.Sprintf("%s/.kubefirst/tools/helm", config.HomePath)

	config.TerraformVersion = "1.0.11"

	// todo adopt latest helmVersion := "v3.9.0"
	config.HelmVersion = "v3.2.1"

	config.KubefirstVersion = "0.1.1"

	config.InstallerEmail = "kubefirst-bot@kubefirst.com"

	return &config
}