package etcdutil

import (
	"crypto/tls"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/etcd/pkg/transport"
)

const (
	CliCertFile = "etcd-client.crt"
	CliKeyFile  = "etcd-client.key"
	CliCAFile   = "etcd-client-ca.crt"
)

func NewTLSConfig(certData, keyData, caData []byte) (*tls.Config, error) {
	dir, err := ioutil.TempDir("", "etcd-operator-cluster-tls")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	certFile, err := writeFile(dir, CliCertFile, certData)
	if err != nil {
		return nil, err
	}
	keyFile, err := writeFile(dir, CliKeyFile, keyData)
	if err != nil {
		return nil, err
	}
	caFile, err := writeFile(dir, CliCAFile, caData)
	if err != nil {
		return nil, err
	}

	tlsInfo := transport.TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: caFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func writeFile(dir, file string, data []byte) (string, error) {
	p := filepath.Join(dir, file)
	return p, ioutil.WriteFile(p, data, 0600)
}
