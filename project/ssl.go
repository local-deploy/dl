package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/cert"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

type tlsStruct struct {
	TSL certStruct `yaml:"tls"`
}

type certStruct struct {
	Certificates []certFileStruct `yaml:"certificates"`
}

type certFileStruct struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

// CreateCert create a certificate and key for the project
func CreateCert() {
	certutilPath, err := utils.CertutilPath()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	c := &cert.Cert{
		CertutilPath:  certutilPath,
		CaFileName:    cert.CaRootName,
		CaFileKeyName: cert.CaRootKeyName,
		CaPath:        utils.CertDir(),
	}

	err = c.LoadCA()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	// ~/.config/dl/certs/site
	certDir := filepath.Join(utils.CertDir(), Env.GetString("NETWORK_NAME"))
	_ = utils.CreateDirectory(certDir)

	err = c.MakeCert(certHosts(Domains), Env.GetString("NETWORK_NAME"))
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
	}

	tls := tlsStruct{
		TSL: certStruct{
			Certificates: []certFileStruct{
				{
					CertFile: "/certs/" + Env.GetString("NETWORK_NAME") + "/cert.pem",
					KeyFile:  "/certs/" + Env.GetString("NETWORK_NAME") + "/key.pem",
				},
			}}}

	yamlData, err := yaml.Marshal(&tls)
	if err != nil {
		fmt.Println(err)
		return
	}

	// ~/.config/dl/certs/conf/site.localhost.yaml
	err = os.WriteFile(filepath.Join(utils.CertDir(), "conf", Env.GetString("NETWORK_NAME")+".yaml"), yamlData, 0600)
	if err != nil {
		pterm.FgRed.Printfln("failed to create config certificate file: %s", err)
	}
}

// certHosts SAN covering the local and the nip.io domain of every project domain
func certHosts(domains []DomainMapping) []string {
	hosts := make([]string, 0, len(domains)*2)
	for _, domain := range domains {
		hosts = append(hosts, domain.LocalDomain, domain.NipDomain)
	}

	return hosts
}
