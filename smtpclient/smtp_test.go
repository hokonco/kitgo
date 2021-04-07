package smtpclient_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/mail"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/smtpclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_smtp(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	ctrl := gomock.NewController(t)
	any := gomock.Any()
	ctx := context.TODO()

	from, to :=
		mail.Address{Address: "mail-1@example.com"},
		mail.Address{Address: "mail-2@example.com"}

	t.Run("ReverseLookup", func(t *testing.T) {
		t.Parallel()
		smtpCli, mock := smtpclient.Test()
		times := 7

		t.Run("err validate", func(t *testing.T) {
			Expect(smtpCli.ReverseLookup(ctx, to, to)).To(HaveOccurred())
			Expect(smtpCli.ReverseLookup(ctx, mail.Address{}, to)).To(HaveOccurred())
			Expect(smtpCli.ReverseLookup(ctx, from, mail.Address{})).To(HaveOccurred())
		})

		netConn := kitgo.NewMockNetConn(ctrl)
		netConn.EXPECT().Read(any).Return(20, io.EOF)
		netConn.EXPECT().Close()

		t.Run("err lookupMX", func(t *testing.T) {
			netResolver := kitgo.NewMockNetResolver(ctrl)
			netResolver.EXPECT().LookupMX(ctx, "example.com").Return(nil, fmt.Errorf("error"))
			mock.WithNetResolver(netResolver)
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())
		})

		netResolver := kitgo.NewMockNetResolver(ctrl)
		netResolver.EXPECT().LookupMX(ctx, "example.com").Return([]*net.MX{{}}, nil).Times(times)
		mock.WithNetResolver(netResolver)

		t.Run("err dialContext", func(t *testing.T) {
			netDialer := kitgo.NewMockNetDialer(ctrl)
			netDialer.EXPECT().DialContext(ctx, "tcp", "example.com:25").Return(nil, fmt.Errorf("error"))
			mock.WithNetDialer(netDialer)
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())
		})

		netDialer := kitgo.NewMockNetDialer(ctrl)
		netDialer.EXPECT().DialContext(ctx, "tcp", "example.com:25").Return(netConn, nil).Times(times - 1)
		mock.WithNetDialer(netDialer)

		Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())

		t.Run("err client", func(t *testing.T) {
			mock.WithNewSmtpClient(nil, fmt.Errorf("error"))
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())
		})
		t.Run("err hello", func(t *testing.T) {
			smtpClient := kitgo.NewMockSmtpClient(ctrl)
			smtpClient.EXPECT().Hello("example.com").Return(fmt.Errorf("error"))
			smtpClient.EXPECT().Close()
			mock.WithNewSmtpClient(smtpClient, nil)
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())
		})
		t.Run("err mail", func(t *testing.T) {
			smtpClient := kitgo.NewMockSmtpClient(ctrl)
			smtpClient.EXPECT().Hello("example.com")
			smtpClient.EXPECT().Mail("mail-1@example.com").Return(fmt.Errorf("error"))
			smtpClient.EXPECT().Close()
			mock.WithNewSmtpClient(smtpClient, nil)
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())
		})
		t.Run("err rcpt", func(t *testing.T) {
			smtpClient := kitgo.NewMockSmtpClient(ctrl)
			smtpClient.EXPECT().Hello("example.com")
			smtpClient.EXPECT().Mail("mail-1@example.com")
			smtpClient.EXPECT().Rcpt("mail-2@example.com").Return(fmt.Errorf("error"))
			smtpClient.EXPECT().Close()
			mock.WithNewSmtpClient(smtpClient, nil)
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(HaveOccurred())
		})
		t.Run("ok", func(t *testing.T) {
			smtpClient := kitgo.NewMockSmtpClient(ctrl)
			smtpClient.EXPECT().Hello("example.com")
			smtpClient.EXPECT().Mail("mail-1@example.com")
			smtpClient.EXPECT().Rcpt("mail-2@example.com")
			smtpClient.EXPECT().Close()
			mock.WithNewSmtpClient(smtpClient, nil)
			Expect(smtpCli.ReverseLookup(ctx, from, to)).To(BeNil())
		})

	})

	t.Run("SendMail", func(t *testing.T) {
		t.Parallel()
		smtpCli, mock := smtpclient.Test()
		_ = mock

		Expect(smtpCli.SendMail(ctx, "", to, to, "", nil, nil)).To(HaveOccurred())
		Expect(smtpCli.SendMail(ctx, "", from, to, "", nil, nil)).To(HaveOccurred())
		Expect(smtpCli.SendMail(ctx, "id", from, to, "subject", []byte("plain"), []byte("html"))).To(HaveOccurred())
		mock.WithSendMail(fmt.Errorf("error"))
		Expect(smtpCli.SendMail(ctx, "id", from, to, "subject", []byte("plain"), []byte("html"))).To(HaveOccurred())
		mock.WithSendMail(nil)
		Expect(smtpCli.SendMail(ctx, "id", from, to, "subject", []byte("plain"), []byte("html"))).To(BeNil())
	})
}
