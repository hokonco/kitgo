package cryptoclient_test

import (
	"bytes"
	"crypto/tls"
	"os"
	"strings"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/cryptoclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_crypto(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	cryptoCli := cryptoclient.New()
	t.Run("SHA & base64", func(t *testing.T) {
		t.Parallel()
		password := cryptoCli.Nonce(16)
		Expect(len(password)).To(Equal(16))
		Expect(len(cryptoCli.SHA512(password))).To(Equal(64))
		Expect(cryptoCli.BtoA([]byte("p45$sW0rd"))).To(Equal([]byte("cDQ1JHNXMHJk")))
		Expect(cryptoCli.AtoB([]byte("cDQ1JHNXMHJk"))).To(Equal([]byte("p45$sW0rd")))
		Expect(cryptoCli.SHA512(password)).To(Equal(cryptoCli.SHA512(password)))
	})
	t.Run("BCrypt", func(t *testing.T) {
		t.Parallel()
		password := cryptoCli.Nonce(16)
		Expect(cryptoCli.BCryptCompareHashAndPassword(cryptoCli.BCrypt(password), password)).To(BeTrue())
	})
	t.Run("SCrypt", func(t *testing.T) {
		t.Parallel()
		password := cryptoCli.Nonce(16)
		Expect(cryptoCli.SCryptCompareHashAndPassword(cryptoCli.SCrypt(password), password)).To(BeTrue())
	})
	t.Run("Argon2I", func(t *testing.T) {
		t.Parallel()
		password := cryptoCli.Nonce(16)
		Expect(cryptoCli.Argon2ICompareHashAndPassword(cryptoCli.Argon2I(password), password)).To(BeTrue())
	})
	t.Run("Argon2ID", func(t *testing.T) {
		t.Parallel()
		password := cryptoCli.Nonce(16)
		Expect(cryptoCli.Argon2IDCompareHashAndPassword(cryptoCli.Argon2ID(password), password)).To(BeTrue())
	})
	t.Run("tls config", func(t *testing.T) {
		t.Parallel()
		conf := cryptoCli.NewTLSConfig("", "", "")
		Expect(conf).NotTo(BeNil())
		cert, err := conf.GetCertificate(&tls.ClientHelloInfo{})
		Expect(err).NotTo(BeNil())
		Expect(cert).To(BeNil())
	})
	t.Run("aesgcm", func(t *testing.T) {
		t.Parallel()
		a := cryptoCli.NewAESGCM([]byte(`random-key`))
		Expect(a).NotTo(BeNil())
		msg := []byte("test valid")
		dec, err := a.Open(a.Seal(msg))
		Expect(err).To(BeNil())
		Expect(dec).To(Equal(msg))
	})
	t.Run("NaCl", func(t *testing.T) {
		t.Parallel()
		password := cryptoCli.Nonce(16)
		nacl := cryptoCli.NewNaCl()
		Expect(nacl).NotTo(BeNil())
		pub, key, err := nacl.SignKeyPair()
		Expect(err).To(BeNil())
		hash := nacl.AuthSum(password, pub)
		Expect(hash).NotTo(BeNil())
		ok := nacl.AuthVerify(hash[:], password, pub)
		Expect(ok).To(BeTrue())
		hash = nacl.Sign(password, key)
		Expect(hash).NotTo(BeNil())
		hash, ok = nacl.SignOpen(hash, pub)
		Expect(ok).To(BeTrue())
		Expect(hash).To(Equal(password))
		hash = nacl.SecretBoxSeal(password, pub)
		hash, ok = nacl.SecretBoxOpen(hash, pub)
		Expect(ok).To(BeTrue())
		Expect(hash).To(Equal(password))
		pub1, key1, err := nacl.BoxKeyPair()
		Expect(err).To(BeNil())
		pub2, key2, err := nacl.BoxKeyPair()
		Expect(err).To(BeNil())
		hash = nacl.BoxSeal(password, pub2, key1)
		hash, ok = nacl.BoxOpen(hash, pub1, key2)
		Expect(ok).To(BeTrue())
		Expect(hash).To(Equal(password))
		hash, err = nacl.BoxSealAnonymous(password, pub1)
		Expect(err).To(BeNil())
		hash, ok = nacl.BoxOpenAnonymous(hash, pub1, key1)
		Expect(ok).To(BeTrue())
		Expect(hash).To(Equal(password))
		shared := nacl.BoxPrecompute(pub1, key2)
		hash = nacl.BoxSealAfterPrecomputation(password, shared)
		hash, ok = nacl.BoxOpenAfterPrecomputation(hash, shared)
		Expect(ok).To(BeTrue())
		Expect(hash).To(Equal(password))
	})
	t.Run("NewRSA", func(t *testing.T) {
		t.Parallel()
		c_RSA, err := cryptoCli.NewRSA(1024)
		Expect(err).To(BeNil())
		Expect(c_RSA).NotTo(BeNil())
	})
	t.Run("NewECDSA", func(t *testing.T) {
		t.Parallel()
		c_ECDSA, err := cryptoCli.NewECDSA(nil)
		Expect(err).To(BeNil())
		Expect(c_ECDSA).NotTo(BeNil())
	})
	t.Run("rsa load", func(t *testing.T) {
		t.Parallel()
		cRSA, err := cryptoCli.ReadRSA(strings.NewReader(``), strings.NewReader(``))
		Expect(err).NotTo(BeNil())
		Expect(cRSA).NotTo(BeNil())

		cRSA, err = cryptoCli.ReadRSA(
			strings.NewReader(`
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`),
			strings.NewReader(`
-----BEGIN RSA PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAxfH1d0Bo6it+cssyCEmteuvHq+eORiS/7AwaBupCTFqKEYXqMWdKSDKZnVSE3SE13ZtTMko2+A3rC4S5oexn
cHYC4Xz751twwhtgE8v5lnDO7BLCghob6hqWkfTmGfniB7pqXj20M1Dtn62isE4AOmLgvLDWU54zbdsQw8OrOI1UxFPWZsZRpRADrwLBh1bKcExXnDgN8WjFudXOJBJh
LKD1sqwUVjfAIqolvDr0l8uEgJ7+yJ0Uaqfqq0QXb1HJyA3OPOoibzO/FjIdjgOpRp7vOaX0KxRHV96jtVe7pyBUbQQoRouwx2AcKNP/YAemXVi/Y0Smyt4olHdESQA+
GlHEas9pNWHHVxozutDHfFBmCOwT92I+HvkfVn6DjjKqpv4264MvY6KWnE4q8qAVF1t1ajntWox8SjDPygqcV6A72S2D2ah2nZU5qMtC7QaTJoZmgadUU6xVPVSKTH9s
T/Yb0/IWKgzg9/ki/ROwvfhhfZgAiRlpxpwRb25Apdt6OYRx4c0CyhxsryGlbqbcU6CxQx02UcsIcq2g7QoEE76px0Q0gQ2jcNrFUUNRwpfrEU8g6ODrnaKXo0rBhR/I
vjwfNS9K2LxA0T2MPrIpKpHaZ+gqGdO5nta7E0jXUic+6VPKpfjIavfueS1LetO7Oj1PVre+3kkWgBjUxVdw20cCAwEAAQ==
-----END RSA PUBLIC KEY-----`),
		)
		Expect(err).To(BeNil())
		Expect(cRSA).NotTo(BeNil())

		err = cRSA.Validate()
		Expect(err).To(BeNil())

		key, pub := &bytes.Buffer{}, &bytes.Buffer{}
		err = cRSA.Write(key, pub)
		Expect(err).To(BeNil())

	})
	t.Run("ecdsa load", func(t *testing.T) {
		t.Parallel()
		cECDSA, err := cryptoCli.ReadECDSA(strings.NewReader(``), strings.NewReader(``))
		Expect(err).NotTo(BeNil())
		Expect(cECDSA).NotTo(BeNil())

		cECDSA, err = cryptoCli.ReadECDSA(
			strings.NewReader(`
-----BEGIN EC PRIVATE KEY-----
MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDAIjPGU225IHg7/7Y6cL15OsyZBeHbjGV6SBUMAFT5tXK7xGfSDXzs+M1PHavIA2LOhZANiAATpMbEBO7em
IiMemTnQeadqE2X+fX/ufdI3aidSmKvrM11gtq9JZSze1qkaWqLQOzQbHAYklhCu0iy6dRSDoj5MzIgnsTCxVKSrzZmbDkwlAwZLEw2LKn2hDmNUJRGSCvI=
-----END EC PRIVATE KEY-----`),
			strings.NewReader(`
-----BEGIN EC PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE6TGxATu3piIjHpk50HmnahNl/n1/7n3SN2onUpir6zNdYLavSWUs3tapGlqi0Ds0GxwGJJYQrtIsunUUg6I+TMyIJ7EwsVSk
q82Zmw5MJQMGSxMNiyp9oQ5jVCURkgry
-----END EC PUBLIC KEY-----`),
		)
		Expect(err).To(BeNil())
		Expect(cECDSA).NotTo(BeNil())

		key, pub := &bytes.Buffer{}, &bytes.Buffer{}
		err = cECDSA.Write(key, pub)
		Expect(err).To(BeNil())
	})
}
