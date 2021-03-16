//go:generate mockgen -package=kitgo -source=kit_mock.go -destination=kit_gen.go
package kitgo

import (
	"context"
	"crypto/tls"
	"io"
	"io/fs"
	"net"
	"net/smtp"
)

var (
	_ NetDialer   = (*net.Dialer)(nil)
	_ NetResolver = (*net.Resolver)(nil)
	_ SmtpClient  = (*smtp.Client)(nil)
)

type FsFileInfo interface{ fs.FileInfo }
type NetAddr interface{ net.Addr }
type NetConn interface{ net.Conn }
type NetDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	Dial(network, address string) (net.Conn, error)
}
type NetResolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupPort(ctx context.Context, network, service string) (port int, err error)
	LookupCNAME(ctx context.Context, host string) (cname string, err error)
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupAddr(ctx context.Context, addr string) (names []string, err error)
}
type SmtpClient interface {
	Close() error
	Hello(localName string) error
	StartTLS(config *tls.Config) error
	TLSConnectionState() (state tls.ConnectionState, ok bool)
	Verify(addr string) error
	Auth(a smtp.Auth) error
	Mail(from string) error
	Rcpt(to string) error
	Data() (io.WriteCloser, error)
	Extension(ext string) (bool, string)
	Reset() error
	Noop() error
	Quit() error
}
