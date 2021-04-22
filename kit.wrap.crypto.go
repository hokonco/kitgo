package kitgo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/nacl/auth"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/nacl/sign"
	"golang.org/x/crypto/scrypt"
)

var Crypto crypto_

type crypto_ struct{}

func (crypto_) New() *CryptoWrapper { return &CryptoWrapper{} }

type CryptoWrapper struct{}

func (*CryptoWrapper) SHA256(b []byte) []byte { cs := sha256.Sum256(b); return cs[:] }

func (*CryptoWrapper) SHA512(b []byte) []byte { cs := sha512.Sum512(b); return cs[:] }

func (*CryptoWrapper) Nonce(length int) []byte {
	nonce := make([]byte, length)
	_, _ = io.ReadFull(rand.Reader, nonce)
	return nonce
}

func (*CryptoWrapper) NewNaCl() *NaCl { return &NaCl{} }

func (*CryptoWrapper) NewAESGCM(password []byte) *AESGCM {
	aesgcm := &AESGCM{}
	if block, err := aes.NewCipher(new(CryptoWrapper).SHA256(password)); err == nil && block != nil {
		aesgcm.AEAD, _ = cipher.NewGCMWithNonceSize(block, block.BlockSize())
	}
	return aesgcm
}
func (*CryptoWrapper) NewRSA(bits int) (*RSA, error) {
	if bits != 4096 && bits != 2048 {
		bits = 2048
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	return &RSA{key}, err
}
func (*CryptoWrapper) ReadRSA(key, pub io.Reader) (*RSA, error) {
	pk := (*rsa.PrivateKey)(nil)
	if kk, err := read(key); kk != nil && err == nil {
		if k, ok := kk.(*rsa.PrivateKey); ok {
			pk = k
		}
	}
	if pp, err := read(pub); pp != nil && err == nil {
		if p, ok := pp.(*rsa.PublicKey); ok && pk != nil {
			pk.PublicKey = *p
		}
	}
	return &RSA{pk}, validateRSA(pk)
}
func (*CryptoWrapper) NewECDSA(curv elliptic.Curve) (*ECDSA, error) {
	if curv == nil {
		curv = elliptic.P384()
	}
	key, err := ecdsa.GenerateKey(curv, rand.Reader)
	return &ECDSA{key}, err
}
func (*CryptoWrapper) ReadECDSA(key, pub io.Reader) (*ECDSA, error) {
	pk := (*ecdsa.PrivateKey)(nil)
	if kk, err := read(key); kk != nil && err == nil {
		if k, ok := kk.(*ecdsa.PrivateKey); ok {
			pk = k
		}
	}
	if pp, err := read(pub); pp != nil && err == nil {
		if p, ok := pp.(*ecdsa.PublicKey); ok && pk != nil {
			pk.PublicKey = *p
		}
	}
	return &ECDSA{pk}, validateECDSA(pk)
}
func (*CryptoWrapper) NewTLSConfig(certFile, keyFile, dirCache string, hostWhitelist ...string) *tls.Config {
	mgr := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(dirCache),
		HostPolicy: autocert.HostWhitelist(hostWhitelist...),
	}
	conf := mgr.TLSConfig()
	conf.GetCertificate = func(hello *tls.ClientHelloInfo) (crt *tls.Certificate, err error) {
		var crt_ tls.Certificate
		crt_, err = tls.LoadX509KeyPair(certFile, keyFile)
		crt = &crt_
		if err != nil {
			crt, err = mgr.GetCertificate(hello)
		}
		return
	}
	return conf
}
func (*CryptoWrapper) BCrypt(password []byte) []byte {
	hash, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	return hash
}
func (*CryptoWrapper) BCryptCompareHashAndPassword(hashed, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashed, password) == nil
}
func (*CryptoWrapper) SCrypt(password []byte) []byte {
	n, r, p, keyLen := 32768, 8, 1, 32
	salt := Crypto.New().Nonce(16)
	hash, _ := scrypt.Key(password, salt, n, r, p, keyLen)
	btoa := Base64.RawStd().BtoA
	b64salt, b64hash := btoa(salt), btoa(hash)
	f := fmt.Sprintf("$scrypt$ln=%d,r=%d,p=%d$%s$%s", n, r, p, b64salt, b64hash)
	return []byte(f)
}
func (*CryptoWrapper) SCryptCompareHashAndPassword(hashed, password []byte) bool {
	n, r, p := 32768, 8, 1
	parts := bytes.Split(hashed, []byte("$"))
	fmt.Sscanf(string(parts[2]), "ln=%d,r=%d,p=%d", &n, &r, &p)
	atob := Base64.RawStd().AtoB
	salt, hash := atob(parts[3]), atob(parts[4])
	z, _ := scrypt.Key(password, salt, n, r, p, len(hash))
	return subtle.ConstantTimeCompare(hash, z) == 1
}
func (*CryptoWrapper) Argon2I(password []byte) []byte {
	t, m, p := uint32(3), uint32(32*1024), uint8(4)
	salt := Crypto.New().Nonce(16)
	hash := argon2.Key(password, salt, t, m, p, 32)
	btoa := Base64.RawStd().BtoA
	b64salt, b64hash := btoa(salt), btoa(hash)
	f := fmt.Sprintf("$argon2i$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, m, t, p, b64salt, b64hash)
	return []byte(f)
}
func (*CryptoWrapper) Argon2ICompareHashAndPassword(hashed, password []byte) bool {
	t, m, p := uint32(0), uint32(0), uint8(0)
	parts := bytes.Split(hashed, []byte("$"))
	fmt.Sscanf(string(parts[3]), "m=%d,t=%d,p=%d", &m, &t, &p)
	atob := Base64.RawStd().AtoB
	salt, hash := atob(parts[4]), atob(parts[5])
	z := argon2.Key(password, salt, t, m, p, uint32(len(hash)))
	return subtle.ConstantTimeCompare(hash, z) == 1
}
func (*CryptoWrapper) Argon2ID(password []byte) []byte {
	t, m, p := uint32(1), uint32(64*1024), uint8(4)
	salt := Crypto.New().Nonce(16)
	hash := argon2.IDKey(password, salt, t, m, p, 32)
	btoa := Base64.RawStd().BtoA
	b64salt, b64hash := btoa(salt), btoa(hash)
	f := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, m, t, p, b64salt, b64hash)
	return []byte(f)
}
func (*CryptoWrapper) Argon2IDCompareHashAndPassword(hashed, password []byte) bool {
	t, m, p := uint32(0), uint32(0), uint8(0)
	parts := bytes.Split(hashed, []byte("$"))
	fmt.Sscanf(string(parts[3]), "m=%d,t=%d,p=%d", &m, &t, &p)
	atob := Base64.RawStd().AtoB
	salt, hash := atob(parts[4]), atob(parts[5])
	z := argon2.IDKey(password, salt, t, m, p, uint32(len(hash)))
	return subtle.ConstantTimeCompare(hash, z) == 1
}

type AESGCM struct{ cipher.AEAD }

func (x *AESGCM) Open(ciphertext []byte) (plaintext []byte, err error) {
	if atob := Base64.RawStd().AtoB; x.AEAD != nil {
		ciphertext = atob(ciphertext)
		plaintext, err = x.AEAD.Open(nil, ciphertext[:x.AEAD.NonceSize()], ciphertext[x.AEAD.NonceSize():], nil)
	}
	return
}
func (x *AESGCM) Seal(plaintext []byte) (ciphertext []byte) {
	if btoa := Base64.RawStd().BtoA; x.AEAD != nil {
		salt := Crypto.New().Nonce(x.AEAD.NonceSize())
		ciphertext = btoa(x.AEAD.Seal(salt, salt, plaintext, nil))
	}
	return
}

type RSA struct{ *rsa.PrivateKey }

func (x *RSA) Write(key, pub io.Writer) (err error) {
	err = validateRSA(x.PrivateKey)
	if err == nil {
		err = write(pub, "RSA PUBLIC KEY", &x.PublicKey)
	}
	if err == nil {
		err = write(key, "RSA PRIVATE KEY", x.PrivateKey)
	}
	return
}
func (x *RSA) EncryptOAEP(msg, label []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, &x.PublicKey, msg, label)
}
func (x *RSA) DecryptOAEP(ciphertext, label []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, x.PrivateKey, ciphertext, label)
}
func (x *RSA) EncryptPKCS1v15(msg []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, &x.PublicKey, msg)
}
func (x *RSA) DecryptPKCS1v15(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, x.PrivateKey, ciphertext)
}

type ECDSA struct{ *ecdsa.PrivateKey }

func (x *ECDSA) Write(key, pub io.Writer) (err error) {
	err = validateECDSA(x.PrivateKey)
	if err == nil {
		err = write(pub, "EC PUBLIC KEY", &x.PublicKey)
	}
	if err == nil {
		err = write(key, "EC PRIVATE KEY", x.PrivateKey)
	}
	return
}
func (x *ECDSA) SignASN1(h []byte) ([]byte, error) {
	fmt.Println("1", x)
	fmt.Println("2", x.PrivateKey)
	fmt.Println("3", h)
	return ecdsa.SignASN1(rand.Reader, x.PrivateKey, h)
}
func (x *ECDSA) VerifyASN1(h, sig []byte) bool {
	return ecdsa.VerifyASN1(&x.PublicKey, h, sig)
}
func (x *ECDSA) Sign(h []byte) (r, s *big.Int, err error) {
	return ecdsa.Sign(rand.Reader, x.PrivateKey, h)
}
func (x *ECDSA) Verify(h []byte, r, s *big.Int) bool {
	return ecdsa.Verify(&x.PublicKey, h, r, s)
}

type NaCl struct{}

func (*NaCl) AuthSum(message []byte, secretKey *[32]byte) []byte {
	hash := auth.Sum(message, secretKey)
	return hash[:]
}
func (*NaCl) AuthVerify(digest []byte, message []byte, secretKey *[32]byte) bool {
	return auth.Verify(digest, message, secretKey)
}
func (*NaCl) BoxKeyPair() (publicKey, privateKey *[32]byte, err error) {
	return box.GenerateKey(rand.Reader)
}
func (*NaCl) BoxPrecompute(peersPublicKey, privateKey *[32]byte) (sharedKey *[32]byte) {
	sharedKey = &[32]byte{}
	box.Precompute(sharedKey, peersPublicKey, privateKey)
	return sharedKey
}
func (*NaCl) BoxOpen(box_ []byte, peersPublicKey, privateKey *[32]byte) ([]byte, bool) {
	n := [24]byte{}
	copy(n[:], box_)
	return box.Open(nil, box_[24:], &n, peersPublicKey, privateKey)
}
func (*NaCl) BoxOpenAfterPrecomputation(box_ []byte, sharedKey *[32]byte) ([]byte, bool) {
	n := [24]byte{}
	copy(n[:], box_)
	return box.OpenAfterPrecomputation(nil, box_[24:], &n, sharedKey)
}
func (*NaCl) BoxOpenAnonymous(box_ []byte, publicKey, privateKey *[32]byte) (message []byte, ok bool) {
	return box.OpenAnonymous(nil, box_, publicKey, privateKey)
}
func (*NaCl) BoxSeal(message []byte, peersPublicKey, privateKey *[32]byte) []byte {
	n := [24]byte{}
	salt := Crypto.New().Nonce(24)
	copy(n[:], salt)
	return box.Seal(salt, message, &n, peersPublicKey, privateKey)
}
func (*NaCl) BoxSealAfterPrecomputation(message []byte, sharedKey *[32]byte) []byte {
	n := [24]byte{}
	salt := Crypto.New().Nonce(24)
	copy(n[:], salt)
	return box.SealAfterPrecomputation(salt, message, &n, sharedKey)
}
func (*NaCl) BoxSealAnonymous(message []byte, recipient *[32]byte) ([]byte, error) {
	return box.SealAnonymous(nil, message, recipient, rand.Reader)
}
func (*NaCl) SecretBoxOpen(box_ []byte, key *[32]byte) ([]byte, bool) {
	salt := [24]byte{}
	copy(salt[:], box_)
	return secretbox.Open(nil, box_[24:], &salt, key)
}
func (*NaCl) SecretBoxSeal(message []byte, key *[32]byte) []byte {
	salt := [24]byte{}
	copy(salt[:], Crypto.New().Nonce(24))
	return secretbox.Seal(salt[:], message, &salt, key)
}
func (*NaCl) SignKeyPair() (publicKey *[32]byte, privateKey *[64]byte, err error) {
	return sign.GenerateKey(rand.Reader)
}
func (*NaCl) SignOpen(signedMessage []byte, publicKey *[32]byte) ([]byte, bool) {
	return sign.Open(nil, signedMessage, publicKey)
}
func (*NaCl) Sign(message []byte, privateKey *[64]byte) []byte {
	return sign.Sign(nil, message, privateKey)
}

// func nonce(length int) (nonce []byte) {
// 	nonce = make([]byte, length)
// 	_, _ = io.ReadFull(rand.Reader, nonce)
// 	return nonce
// }

// var b64 = base64.RawStdEncoding

// func btoa(b []byte) (enc []byte) {
// 	enc = make([]byte, b64.EncodedLen(len(b)))
// 	b64.Encode(enc, b)
// 	return
// }
// func atob(b []byte) (dec []byte) {
// 	dec = make([]byte, b64.DecodedLen(len(b)))
// 	_, _ = b64.Decode(dec, b)
// 	return
// }
func read(r io.Reader) (k interface{}, err error) {
	buf := &bytes.Buffer{}
	if _, err = io.Copy(buf, io.LimitReader(r, 1e9)); err == nil {
		p, rb := pem.Decode(buf.Bytes())
		err = fmt.Errorf("invalid pem:[%v] rest:[%s]", p, rb)
		if p != nil {
			k, err = x509.ParsePKCS8PrivateKey(p.Bytes)
			if err != nil || k == nil {
				k, err = x509.ParsePKIXPublicKey(p.Bytes)
			}
		}
	}
	return k, err
}
func write(w io.Writer, kt string, k interface{}) (err error) {
	var b []byte
	if b, err = x509.MarshalPKCS8PrivateKey(k); err != nil || len(b) < 1 {
		b, err = x509.MarshalPKIXPublicKey(k)
	}
	if err == nil {
		err = pem.Encode(w, &pem.Block{Type: kt, Bytes: b})
	}
	return
}
func validateRSA(pk *rsa.PrivateKey) (err error) {
	if pk == nil {
		pk = &rsa.PrivateKey{PublicKey: rsa.PublicKey{}}
	}
	err = pk.Validate()
	msg, lbl := []byte("test valid"), []byte("label 123")
	x, e1, d1, e2, d2 := &RSA{pk}, []byte{}, []byte{}, []byte{}, []byte{}
	if err == nil {
		e1, err = x.EncryptOAEP(msg, lbl)
	}
	if err == nil {
		d1, err = x.DecryptOAEP(e1, lbl)
	}
	if err == nil {
		e2, err = x.EncryptPKCS1v15(msg)
	}
	if err == nil {
		d2, err = x.DecryptPKCS1v15(e2)
	}
	cmp := func(a, b []byte) bool { return string(a) == string(b) }
	if err == nil {
		err = fmt.Errorf("mismatched: msg:[%s] d1:[%s] d2:[%s]", msg, d1, d2)
	}
	if cmp(msg, d1) && cmp(msg, d2) {
		err = nil
	}
	return
}
func validateECDSA(pk *ecdsa.PrivateKey) (err error) {
	if pk == nil {
		pk = &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{}}
	}
	onCurve, x, h := true, &ECDSA{pk}, sha256.Sum256([]byte("test valid"))
	onCurve = onCurve && x.PrivateKey.Curve != nil
	onCurve = onCurve && x.PublicKey.Curve != nil
	onCurve = onCurve && x.PrivateKey.IsOnCurve(x.PrivateKey.X, x.PrivateKey.Y)
	onCurve = onCurve && x.PublicKey.IsOnCurve(x.PublicKey.X, x.PublicKey.Y)
	if !onCurve {
		err = fmt.Errorf("pub not on curve: key_x:[%v] key_y:[%v] pub_x:[%v] pub_y:[%v]",
			x.PrivateKey.X, x.PrivateKey.Y,
			x.PublicKey.X, x.PublicKey.Y,
		)
	}
	sig := []byte{}
	if err == nil {
		sig, err = x.SignASN1(h[:])
	}
	r, s := (*big.Int)(nil), (*big.Int)(nil)
	if err == nil {
		r, s, err = x.Sign(h[:])
	}
	if err == nil {
		err = fmt.Errorf("mismatch: sig:[%s] r:[%v] s:[%v]", sig, r, s)
		if x.VerifyASN1(h[:], sig) && x.Verify(h[:], r, s) {
			err = nil
		}
	}
	return
}
