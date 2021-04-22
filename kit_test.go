package kitgo_test

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_mock(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	ctrl := gomock.NewController(t)
	any := gomock.Any()
	ctx := context.TODO()
	ti := time.Time{}

	t.Run("mockFsFileInfo", func(t *testing.T) {
		mockFsFileInfo := kitgo.NewMockFsFileInfo(ctrl)
		mockFsFileInfo.EXPECT().Name().Return("name")
		mockFsFileInfo.EXPECT().Size().Return(int64(100))
		mockFsFileInfo.EXPECT().Mode().Return(os.ModeAppend)
		mockFsFileInfo.EXPECT().ModTime().Return(time.Time{})
		mockFsFileInfo.EXPECT().IsDir().Return(false)
		mockFsFileInfo.EXPECT().Sys().Return("<nil>")
		Expect(mockFsFileInfo.Name()).To(Equal("name"))
		Expect(mockFsFileInfo.Size()).To(Equal(int64(100)))
		Expect(mockFsFileInfo.Mode()).To(Equal(os.ModeAppend))
		Expect(mockFsFileInfo.ModTime()).To(Equal(time.Time{}))
		Expect(mockFsFileInfo.IsDir()).To(Equal(false))
		Expect(mockFsFileInfo.Sys()).To(Equal("<nil>"))
	})
	t.Run("mockNetAddr", func(t *testing.T) {
		mockNetAddr := kitgo.NewMockNetAddr(ctrl)
		mockNetAddr.EXPECT().Network().Return("network")
		mockNetAddr.EXPECT().String().Return("string")
		Expect(mockNetAddr.Network()).To(Equal("network"))
		Expect(mockNetAddr.String()).To(Equal("string"))
	})
	t.Run("mockNetConn", func(t *testing.T) {
		mockNetConn := kitgo.NewMockNetConn(ctrl)
		mockNetConn.EXPECT().Read(any)
		mockNetConn.EXPECT().Write(any)
		mockNetConn.EXPECT().Close()
		mockNetConn.EXPECT().LocalAddr()
		mockNetConn.EXPECT().RemoteAddr()
		mockNetConn.EXPECT().SetDeadline(any)
		mockNetConn.EXPECT().SetReadDeadline(any)
		mockNetConn.EXPECT().SetWriteDeadline(any)
		_, _ = mockNetConn.Read(nil)
		_, _ = mockNetConn.Write(nil)
		_ = mockNetConn.Close()
		_ = mockNetConn.LocalAddr()
		_ = mockNetConn.RemoteAddr()
		_ = mockNetConn.SetDeadline(ti)
		_ = mockNetConn.SetReadDeadline(ti)
		_ = mockNetConn.SetWriteDeadline(ti)
	})
	t.Run("mockNetDialer", func(t *testing.T) {
		mockNetDialer := kitgo.NewMockNetDialer(ctrl)
		mockNetDialer.EXPECT().DialContext(any, any, any)
		mockNetDialer.EXPECT().Dial(any, any)
		_, _ = mockNetDialer.DialContext(ctx, "", "")
		_, _ = mockNetDialer.Dial("", "")
	})
	t.Run("mockNetResolver", func(t *testing.T) {
		mockNetResolver := kitgo.NewMockNetResolver(ctrl)
		mockNetResolver.EXPECT().LookupHost(any, any)
		mockNetResolver.EXPECT().LookupIPAddr(any, any)
		mockNetResolver.EXPECT().LookupIP(any, any, any)
		mockNetResolver.EXPECT().LookupPort(any, any, any)
		mockNetResolver.EXPECT().LookupCNAME(any, any)
		mockNetResolver.EXPECT().LookupSRV(any, any, any, any)
		mockNetResolver.EXPECT().LookupMX(any, any)
		mockNetResolver.EXPECT().LookupNS(any, any)
		mockNetResolver.EXPECT().LookupTXT(any, any)
		mockNetResolver.EXPECT().LookupAddr(any, any)
		_, _ = mockNetResolver.LookupHost(ctx, "")
		_, _ = mockNetResolver.LookupIPAddr(ctx, "")
		_, _ = mockNetResolver.LookupIP(ctx, "", "")
		_, _ = mockNetResolver.LookupPort(ctx, "", "")
		_, _ = mockNetResolver.LookupCNAME(ctx, "")
		_, _, _ = mockNetResolver.LookupSRV(ctx, "", "", "")
		_, _ = mockNetResolver.LookupMX(ctx, "")
		_, _ = mockNetResolver.LookupNS(ctx, "")
		_, _ = mockNetResolver.LookupTXT(ctx, "")
		_, _ = mockNetResolver.LookupAddr(ctx, "")
	})
	t.Run("mockSmtpClient", func(t *testing.T) {
		mockSmtpClient := kitgo.NewMockSmtpClient(ctrl)
		mockSmtpClient.EXPECT().Close()
		mockSmtpClient.EXPECT().Hello(any)
		mockSmtpClient.EXPECT().StartTLS(any)
		mockSmtpClient.EXPECT().TLSConnectionState()
		mockSmtpClient.EXPECT().Verify(any)
		mockSmtpClient.EXPECT().Auth(any)
		mockSmtpClient.EXPECT().Mail(any)
		mockSmtpClient.EXPECT().Rcpt(any)
		mockSmtpClient.EXPECT().Data()
		mockSmtpClient.EXPECT().Extension(any)
		mockSmtpClient.EXPECT().Reset()
		mockSmtpClient.EXPECT().Noop()
		mockSmtpClient.EXPECT().Quit()

		_ = mockSmtpClient.Close()
		_ = mockSmtpClient.Hello("")
		_ = mockSmtpClient.StartTLS(nil)
		_, _ = mockSmtpClient.TLSConnectionState()
		_ = mockSmtpClient.Verify("")
		_ = mockSmtpClient.Auth(nil)
		_ = mockSmtpClient.Mail("")
		_ = mockSmtpClient.Rcpt("")
		_, _ = mockSmtpClient.Data()
		_, _ = mockSmtpClient.Extension("")
		_ = mockSmtpClient.Reset()
		_ = mockSmtpClient.Noop()
		_ = mockSmtpClient.Quit()
	})
}

func Test_pkg_datatype(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	Expect(string(kitgo.Template.HTML.CSS("*{color:red}"))).To(Equal("*{color:red}"))
	Expect(string(kitgo.Template.HTML.HTML("<p>p</p>"))).To(Equal("<p>p</p>"))
	Expect(string(kitgo.Template.HTML.HTMLAttr("id=id"))).To(Equal("id=id"))
	Expect(string(kitgo.Template.HTML.JS("alert(1)"))).To(Equal("alert(1)"))
	Expect(string(kitgo.Template.HTML.JSStr("\"str\""))).To(Equal("\"str\""))
	Expect(string(kitgo.Template.HTML.URL("http://localhost/"))).To(Equal("http://localhost/"))
	Expect(string(kitgo.Template.HTML.Srcset("/a.jpg,/b.jpg"))).To(Equal("/a.jpg,/b.jpg"))

	d := kitgo.Dict{}
	d.Set("a", 1)
	d.Map(func(k string, v interface{}) { d[k] = fmt.Sprint(v) })
	d.Delete("a")

	l := kitgo.List{}
	l.Add(1)
	l.Map(func(k int, v interface{}) { l[k] = fmt.Sprint(v) })
	l.Delete(0)

	_ = kitgo.ShouldCover(0, 1)
	// _ = kitgo.ParseDuration("", 0)

	go func() { kitgo.ListenToSignal(os.Kill) }()
	os.Kill.Signal()
}

func Test_pkg_panic(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect
	var err error

	kitgo.PanicWhen(err != nil, err)
	Expect(func() { kitgo.PanicWhen(err == nil, err) }).NotTo(Panic())

	err = fmt.Errorf("error")
	defer kitgo.RecoverWith(func(v interface{}) {})
	kitgo.PanicWhen(err != nil, err)
}

func Test_pkg_errors(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	err := fmt.Errorf("error")
	Expect(err).NotTo(BeNil())

	errs := kitgo.NewErrors(err)
	Expect(errs).NotTo(BeNil())
	Expect(errs.Error()).To(Equal(err.Error()))

	errs = errs.Append(err)
	Expect(errs).NotTo(BeNil())
	Expect(errs.Error()).To(Equal(err.Error() + "\n" + err.Error()))

	errs = kitgo.NewErrors(nil, nil, err, fmt.Errorf("error"))
	Expect(errs).NotTo(BeNil())
	Expect(errs.Error()).To(Equal(err.Error() + "\n" + err.Error()))

	b, err := errs.MarshalJSON()
	Expect(err).To(BeNil())
	Expect(b).NotTo(BeNil())
	Expect(b).To(Equal([]byte(`["error","error"]`)))

	errs = kitgo.NewErrors(nil)
	Expect(errs).To(BeNil())
	Expect(len(errs)).To(Equal(0))

	errs = kitgo.NewErrors()
	err = errs.UnmarshalJSON([]byte(`["error","error"]`))
	Expect(err).To(BeNil())
	Expect(len(errs)).To(Equal(2))
}

func Test_pkg_currency(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	value := 1_234_567.89123456789
	tests := []struct {
		expect    string
		parameter kitgo.Currency
	}{
		{"\"ج.م.\u200f ١٬٢٣٤٬٥٦٧٫٨٩\"", kitgo.Currency{Tag: "ar", Value: value}},
		{"\"лв. 1 234 567,89\"", kitgo.Currency{Tag: "bg", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "ca", Value: value, Format: "%[1]s%[2]s"}},
		{"\"￥ 1,234,567.89\"", kitgo.Currency{Tag: "zh-Hans", Value: value}},
		{"\"1 234 567,89 Kč\"", kitgo.Currency{Tag: "cs", Value: value, Format: "%[2]s %[1]s"}},
		{"\"kr. 1.234.567,89\"", kitgo.Currency{Tag: "da", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "de", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "es", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fi", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fr", Value: value, Format: "%[1]s%[2]s"}},
		{"\"₪ 1,234,567.89\"", kitgo.Currency{Tag: "he", Value: value}},
		{"\"Ft 1 234 567,89\"", kitgo.Currency{Tag: "hu", Value: value}},
		{"\"ISK 1.234.568\"", kitgo.Currency{Tag: "is", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "it", Value: value, Format: "%[1]s%[2]s"}},
		{"\"￥ 1,234,568\"", kitgo.Currency{Tag: "ja", Value: value}},
		{"\"₩ 1,234,568\"", kitgo.Currency{Tag: "ko", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "nl", Value: value, Format: "%[1]s%[2]s"}},
		{"\"NOK 1,234,567.89\"", kitgo.Currency{Tag: "no", Value: value}},
		{"\"1 234 567,89 zł\"", kitgo.Currency{Tag: "pl", Value: value, Format: "%[2]s %[1]s"}},
		{"\"R$ 1.234.567,89\"", kitgo.Currency{Tag: "pt", Value: value}},
		{"\"CHF 1’234’567.89\"", kitgo.Currency{Tag: "rm", Value: value}},
		{"\"RON 1.234.567,89\"", kitgo.Currency{Tag: "ro", Value: value}},
		{"\"1 234 567,89 ₽\"", kitgo.Currency{Tag: "ru", Value: value, Format: "%[2]s %[1]s"}},
		{"\"HRK 1.234.567,89\"", kitgo.Currency{Tag: "hr", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "sk", Value: value, Format: "%[1]s%[2]s"}},
		{"\"Lekë 1 234 567,89\"", kitgo.Currency{Tag: "sq", Value: value}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "sv", Value: value}},
		{"\"THB 1,234,567.89\"", kitgo.Currency{Tag: "th", Value: value}},
		{"\"₺ 1.234.567,89\"", kitgo.Currency{Tag: "tr", Value: value}},
		{"\"Rs 1,234,567.89\"", kitgo.Currency{Tag: "ur", Value: value}},
		{"\"Rp 1.234.567,89\"", kitgo.Currency{Tag: "id", Value: value}},
		{"\"₴ 1 234 567,89\"", kitgo.Currency{Tag: "uk", Value: value}},
		{"\"Br 1 234 567,89\"", kitgo.Currency{Tag: "be", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "sl", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "et", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "lv", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "lt", Value: value, Format: "%[1]s%[2]s"}},
		{"\"сом. 1 234 567,89\"", kitgo.Currency{Tag: "tg", Value: value}},
		{"\"ریال ۱٬۲۳۴٬۵۶۷٫۸۹\"", kitgo.Currency{Tag: "fa", Value: value}},
		{"\"₫ 1.234.568\"", kitgo.Currency{Tag: "vi", Value: value}},
		{"\"֏ 1 234 567,89\"", kitgo.Currency{Tag: "hy", Value: value}},
		{"\"₼ 1.234.567,89\"", kitgo.Currency{Tag: "az", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "eu", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "hsb", Value: value, Format: "%[1]s%[2]s"}},
		{"\"ден 1.234.567,89\"", kitgo.Currency{Tag: "mk", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "tn", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "xh", Value: value}},
		{"\"R 1,234,567.89\"", kitgo.Currency{Tag: "zu", Value: value}},
		{"\"R 1 234 567,89\"", kitgo.Currency{Tag: "af", Value: value}},
		{"\"₾ 1 234 567,89\"", kitgo.Currency{Tag: "ka", Value: value}},
		{"\"kr 1.234.567,89\"", kitgo.Currency{Tag: "fo", Value: value}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "hi", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "mt", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "se", Value: value}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "ga", Value: value, Format: "%[1]s%[2]s"}},
		{"\"RM 1,234,567.89\"", kitgo.Currency{Tag: "ms", Value: value}},
		{"\"₸ 1 234 567,89\"", kitgo.Currency{Tag: "kk", Value: value}},
		{"\"сом 1 234 567,89\"", kitgo.Currency{Tag: "ky", Value: value}},
		{"\"TSh 1,234,567.89\"", kitgo.Currency{Tag: "sw", Value: value}},
		{"\"TMT 1 234 567,89\"", kitgo.Currency{Tag: "tk", Value: value}},
		{"\"soʻm 1 234 567,89\"", kitgo.Currency{Tag: "uz", Value: value}},
		{"\"1 234 567,89 ₽\"", kitgo.Currency{Tag: "tt", Value: value, Format: "%[2]s %[1]s"}},
		{"\"৳  ১২,৩৪,৫৬৭.৮৯\"", kitgo.Currency{Tag: "bn", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "pa", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "gu", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "or", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "ta", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "te", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "kn", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "ml", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  ১২,৩৪,৫৬৭.৮৯\"", kitgo.Currency{Tag: "as", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  १२,३४,५६७.८९\"", kitgo.Currency{Tag: "mr", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "sa", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₮ 1,234,567.89\"", kitgo.Currency{Tag: "mn", Value: value}},
		{"\"¥ 1,234,567.89\"", kitgo.Currency{Tag: "bo", Value: value}},
		{"\"£ 1,234,567.89\"", kitgo.Currency{Tag: "cy", Value: value}},
		{"\"៛ 1.234.567,89\"", kitgo.Currency{Tag: "km", Value: value}},
		{"\"₭ 1.234.567,89\"", kitgo.Currency{Tag: "lo", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "gl", Value: value, Format: "%[1]s%[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "kok", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"IQD 1,234,567.891\"", kitgo.Currency{Tag: "syr", Value: value}},
		{"\"රු. 1,234,567.89\"", kitgo.Currency{Tag: "si", Value: value}},
		{"\"CA$ 1,234,567.89\"", kitgo.Currency{Tag: "iu", Value: value}},
		{"\"ብር 1,234,567.89\"", kitgo.Currency{Tag: "am", Value: value}},
		{"\"MAD 1 234 567,89\"", kitgo.Currency{Tag: "tzm", Value: value}},
		{"\"नेरू १,२३४,५६७.८९\"", kitgo.Currency{Tag: "ne", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "fy", Value: value, Format: "%[1]s%[2]s"}},
		{"\"؋ ۱٬۲۳۴٬۵۶۷٫۸۹\"", kitgo.Currency{Tag: "ps", Value: value}},
		{"\"₱ 1,234,567.89\"", kitgo.Currency{Tag: "fil", Value: value}},
		{"\"MVR 1,234,567.89\"", kitgo.Currency{Tag: "dv", Value: value}},
		{"\"₦ 1,234,567.89\"", kitgo.Currency{Tag: "ha", Value: value}},
		{"\"₦ 1,234,567.89\"", kitgo.Currency{Tag: "yo", Value: value}},
		{"\"PEN 1,234,567.89\"", kitgo.Currency{Tag: "quz", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "nso", Value: value}},
		{"\"RUB 1,234,567.89\"", kitgo.Currency{Tag: "ba", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "lb", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr. 1.234.567,89\"", kitgo.Currency{Tag: "kl", Value: value}},
		{"\"₦ 1,234,567.89\"", kitgo.Currency{Tag: "ig", Value: value}},
		{"\"¥ 1,234,567.89\"", kitgo.Currency{Tag: "ii", Value: value}},
		{"\"CLP 1,234,568\"", kitgo.Currency{Tag: "arn", Value: value}},
		{"\"CA$ 1,234,567.89\"", kitgo.Currency{Tag: "moh", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "br", Value: value, Format: "%[1]s%[2]s"}},
		{"\"￥ 1,234,567.89\"", kitgo.Currency{Tag: "ug", Value: value}},
		{"\"NZ$ 1,234,567.89\"", kitgo.Currency{Tag: "mi", Value: value}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "oc", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "co", Value: value, Format: "%[1]s%[2]s"}},
		{"\"CHF 1’234’567.89\"", kitgo.Currency{Tag: "gsw", Value: value}},
		{"\"1 234 567,89 ₽\"", kitgo.Currency{Tag: "sah", Value: value, Format: "%[2]s %[1]s"}},
		//# {"\"", pkg.Currency{Tag: "qut", Value: value}},
		{"\"RF 1.234.568\"", kitgo.Currency{Tag: "rw", Value: value}},
		{"\"CFA 1.234.568\"", kitgo.Currency{Tag: "wo", Value: value}},
		{"\"XXX 1,234,567.89\"", kitgo.Currency{Tag: "prs", Value: value}},
		{"\"£ 1,234,567.89\"", kitgo.Currency{Tag: "gd", Value: value}},

		// else
		{"\"ر.س.\u200f 1,234,567.89\"", kitgo.Currency{Tag: "ar-SA", Value: value}},
		{"\"лв. 1 234 567,89\"", kitgo.Currency{Tag: "bg-BG", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "ca-ES", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "zh-TW", Value: value}},
		{"\"1 234 567,89 Kč\"", kitgo.Currency{Tag: "cs-CZ", Value: value, Format: "%[2]s %[1]s"}},
		{"\"kr. 1.234.567,89\"", kitgo.Currency{Tag: "da-DK", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "de-DE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "el-GR", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-US", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "es-ES", Value: value, Format: "%[1]s%[2]s"}},
		//# {"\"", pkg.Currency{Tag: "es-ES_tradnl", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fi-FI", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fr-FR", Value: value, Format: "%[1]s%[2]s"}},
		{"\"₪ 1,234,567.89\"", kitgo.Currency{Tag: "he-IL", Value: value}},
		{"\"Ft 1 234 567,89\"", kitgo.Currency{Tag: "hu-HU", Value: value}},
		{"\"ISK 1.234.568\"", kitgo.Currency{Tag: "is-IS", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "it-IT", Value: value, Format: "%[1]s%[2]s"}},
		{"\"￥ 1,234,568\"", kitgo.Currency{Tag: "ja-JP", Value: value}},
		{"\"₩ 1,234,568\"", kitgo.Currency{Tag: "ko-KR", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "nl-NL", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "nb-NO", Value: value}},
		{"\"1 234 567,89 zł\"", kitgo.Currency{Tag: "pl-PL", Value: value, Format: "%[2]s %[1]s"}},
		{"\"R$ 1.234.567,89\"", kitgo.Currency{Tag: "pt-BR", Value: value}},
		{"\"CHF 1’234’567.89\"", kitgo.Currency{Tag: "rm-CH", Value: value}},
		{"\"RON 1.234.567,89\"", kitgo.Currency{Tag: "ro-RO", Value: value}},
		{"\"1 234 567,89 ₽\"", kitgo.Currency{Tag: "ru-Ru", Value: value, Format: "%[2]s %[1]s"}},
		{"\"HRK 1.234.567,89\"", kitgo.Currency{Tag: "hr-HR", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "sk-SK", Value: value, Format: "%[1]s%[2]s"}},
		{"\"Lekë 1 234 567,89\"", kitgo.Currency{Tag: "sq-AL", Value: value}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "sv-SE", Value: value}},
		{"\"THB 1,234,567.89\"", kitgo.Currency{Tag: "th-TH", Value: value}},
		{"\"₺ 1.234.567,89\"", kitgo.Currency{Tag: "tr-TR", Value: value}},
		{"\"Rs 1,234,567.89\"", kitgo.Currency{Tag: "ur-PK", Value: value}},
		{"\"Rp 1.234.567,89\"", kitgo.Currency{Tag: "id-ID", Value: value}},
		{"\"₴ 1 234 567,89\"", kitgo.Currency{Tag: "uk-UA", Value: value}},
		{"\"Br 1 234 567,89\"", kitgo.Currency{Tag: "be-BY", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "sl-SI", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "et-EE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "lv-LV", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "lt-LT", Value: value, Format: "%[1]s%[2]s"}},
		{"\"сом. 1 234 567,89\"", kitgo.Currency{Tag: "tg-Cyrl-TJ", Value: value}},
		{"\"ریال 1,234,567.89\"", kitgo.Currency{Tag: "fa-IR", Value: value}},
		{"\"₫ 1.234.568\"", kitgo.Currency{Tag: "vi-VN", Value: value}},
		{"\"֏ 1 234 567,89\"", kitgo.Currency{Tag: "hy-AM", Value: value}},
		{"\"₼ 1.234.567,89\"", kitgo.Currency{Tag: "az-Latn-AZ", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "eu-ES", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "eu-ES", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "wen-DE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"ден 1.234.567,89\"", kitgo.Currency{Tag: "mk-MK", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "st-ZA", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "ts-ZA", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "tn-ZA", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "ven-ZA", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "xh-ZA", Value: value}},
		{"\"R 1,234,567.89\"", kitgo.Currency{Tag: "zu-ZA", Value: value}},
		{"\"R 1 234 567,89\"", kitgo.Currency{Tag: "af-ZA", Value: value}},
		{"\"₾ 1 234 567,89\"", kitgo.Currency{Tag: "ka-GE", Value: value}},
		{"\"kr 1.234.567,89\"", kitgo.Currency{Tag: "fo-FO", Value: value}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "hi-in", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "mt-MT", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "se-NO", Value: value}},
		{"\"RM 1,234,567.89\"", kitgo.Currency{Tag: "ms-MY", Value: value}},
		{"\"₸ 1 234 567,89\"", kitgo.Currency{Tag: "kk-KZ", Value: value}},
		{"\"сом 1 234 567,89\"", kitgo.Currency{Tag: "ky-KG", Value: value}},
		{"\"Ksh 1,234,567.89\"", kitgo.Currency{Tag: "sw-KE", Value: value}},
		{"\"TMT 1 234 567,89\"", kitgo.Currency{Tag: "tk-TM", Value: value}},
		{"\"soʻm 1 234 567,89\"", kitgo.Currency{Tag: "uz-Latn-UZ", Value: value}},
		{"\"1 234 567,89 ₽\"", kitgo.Currency{Tag: "tt-RU", Value: value, Format: "%[2]s %[1]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "bn-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "pa-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "gu-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "or-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "ta-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "te-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "kn-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "ml-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "as-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "mr-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "sa-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₮ 1,234,567.89\"", kitgo.Currency{Tag: "mn-MN", Value: value}},
		{"\"¥ 1,234,567.89\"", kitgo.Currency{Tag: "bo-CN", Value: value}},
		{"\"£ 1,234,567.89\"", kitgo.Currency{Tag: "cy-GB", Value: value}},
		{"\"៛ 1.234.567,89\"", kitgo.Currency{Tag: "km-KH", Value: value}},
		{"\"₭ 1.234.567,89\"", kitgo.Currency{Tag: "lo-LA", Value: value}},
		{"\"K 1,234,567.89\"", kitgo.Currency{Tag: "my-MM", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "gl-ES", Value: value, Format: "%[1]s%[2]s"}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "kok-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  1,234,567.89\"", kitgo.Currency{Tag: "mni", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₹  ١٬٢٣٤٬٥٦٧٫٨٩\"", kitgo.Currency{Tag: "sd-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"SYP 1,234,567.89\"", kitgo.Currency{Tag: "syr-SY", Value: value}},
		{"\"රු. 1,234,567.89\"", kitgo.Currency{Tag: "si-LK", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "chr-US", Value: value}},
		{"\"CA$ 1,234,567.89\"", kitgo.Currency{Tag: "iu-Cans-CA", Value: value}},
		{"\"ብር 1,234,567.89\"", kitgo.Currency{Tag: "am-ET", Value: value}},
		{"\"XXX 1,234,567.89\"", kitgo.Currency{Tag: "tmz", Value: value}},
		{"\"नेरू 1,234,567.89\"", kitgo.Currency{Tag: "ne-NP", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "fy-NL", Value: value, Format: "%[1]s%[2]s"}},
		{"\"؋ 1.234.567,89\"", kitgo.Currency{Tag: "ps-AF", Value: value}},
		{"\"₱ 1,234,567.89\"", kitgo.Currency{Tag: "fil-PH", Value: value}},
		{"\"MVR 1,234,567.89\"", kitgo.Currency{Tag: "dv-MV", Value: value}},
		{"\"NGN 1,234,567.89\"", kitgo.Currency{Tag: "bin-NG", Value: value}},
		{"\"NGN 1,234,567.89\"", kitgo.Currency{Tag: "fuv-NG", Value: value}},
		{"\"₦ 1,234,567.89\"", kitgo.Currency{Tag: "ha-Latn-NG", Value: value}},
		{"\"NGN 1,234,567.89\"", kitgo.Currency{Tag: "ibb-NG", Value: value}},
		{"\"₦ 1,234,567.89\"", kitgo.Currency{Tag: "yo-NG", Value: value}},
		{"\"BOB 1,234,567.89\"", kitgo.Currency{Tag: "quz-BO", Value: value}},
		{"\"ZAR 1,234,567.89\"", kitgo.Currency{Tag: "nso-ZA", Value: value}},
		{"\"RUB 1,234,567.89\"", kitgo.Currency{Tag: "ba-RU", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "lb-LU", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr. 1.234.567,89\"", kitgo.Currency{Tag: "kl-GL", Value: value}},
		{"\"₦ 1,234,567.89\"", kitgo.Currency{Tag: "ig-NG", Value: value}},
		{"\"NGN 1,234,567.89\"", kitgo.Currency{Tag: "kr-NG", Value: value}},
		{"\"ETB 1,234,567.89\"", kitgo.Currency{Tag: "gaz-ET", Value: value}},
		{"\"Nfk 1,234,567.89\"", kitgo.Currency{Tag: "ti-ER", Value: value}},
		{"\"PYG 1,234,568\"", kitgo.Currency{Tag: "gn-PY", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "haw-US", Value: value}},
		{"\"S 1,234,567.89\"", kitgo.Currency{Tag: "so-SO", Value: value}},
		{"\"¥ 1,234,567.89\"", kitgo.Currency{Tag: "ii-CN", Value: value}},
		{"\"XXX 1,234,567.89\"", kitgo.Currency{Tag: "pap-AN", Value: value}},
		{"\"CLP 1,234,568\"", kitgo.Currency{Tag: "arn-CL", Value: value}},
		{"\"CA$ 1,234,567.89\"", kitgo.Currency{Tag: "moh-CA", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "br-FR", Value: value, Format: "%[1]s%[2]s"}},
		{"\"￥ 1,234,567.89\"", kitgo.Currency{Tag: "ug-CN", Value: value}},
		{"\"NZ$ 1,234,567.89\"", kitgo.Currency{Tag: "mi-NZ", Value: value}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "oc-FR", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "co-FR", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1’234’567.89\"", kitgo.Currency{Tag: "gsw-FR", Value: value, Format: "%[1]s%[2]s"}},
		{"\"1 234 567,89 ₽\"", kitgo.Currency{Tag: "sah-RU", Value: value, Format: "%[2]s %[1]s"}},
		//# {"\"", pkg.Currency{Tag: "qut-GT", Value: value}},
		{"\"RF 1.234.568\"", kitgo.Currency{Tag: "rw-RW", Value: value}},
		{"\"CFA 1.234.568\"", kitgo.Currency{Tag: "wo-SN", Value: value}},
		{"\"AFN 1,234,567.89\"", kitgo.Currency{Tag: "prs-AF", Value: value}},
		{"\"MGA 1,234,567.89\"", kitgo.Currency{Tag: "plt-MG", Value: value}},
		{"\"£ 1,234,567.89\"", kitgo.Currency{Tag: "gd-GB", Value: value}},
		{"\"￥ 1,234,567.89\"", kitgo.Currency{Tag: "zh-CN", Value: value}},
		{"\"CHF 1’234’567.89\"", kitgo.Currency{Tag: "de-CH", Value: value}},
		{"\"£ 1,234,567.89\"", kitgo.Currency{Tag: "en-GB", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "es-MX", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fr-BE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"CHF 1’234’567.89\"", kitgo.Currency{Tag: "it-CH", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "nl-BE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "nn-NO", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "pt-PT", Value: value, Format: "%[1]s%[2]s"}},
		{"\"MOP 1.234.567,89\"", kitgo.Currency{Tag: "ro-MO", Value: value}},
		{"\"MOP 1 234 567,89\"", kitgo.Currency{Tag: "ru-MO", Value: value}},
		{"\"XXX 1.234.567,89\"", kitgo.Currency{Tag: "sr-Latn-CS", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "sv-FI", Value: value, Format: "%[1]s%[2]s"}},
		{"\"₹  ۱٬۲۳۴٬۵۶۷٫۸۹\"", kitgo.Currency{Tag: "ur-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"₼ 1.234.567,89\"", kitgo.Currency{Tag: "az-Cyrl-AZ", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "dsb-DE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "se-SE", Value: value}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "ga-IE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1.234.567,89\"", kitgo.Currency{Tag: "ms-BN", Value: value}},
		{"\"сўм 1 234 567,89\"", kitgo.Currency{Tag: "uz-Cyrl-UZ", Value: value}},
		{"\"৳ 12,34,567.89\"", kitgo.Currency{Tag: "bn-BD", Value: value}},
		{"\"ر 1,234,567.89\"", kitgo.Currency{Tag: "pa-PK", Value: value}},
		{"\"CN¥ 1,234,567.89\"", kitgo.Currency{Tag: "mn-Mong-CN", Value: value}},
		{"\"BTN 1,234,567.89\"", kitgo.Currency{Tag: "bo-BT", Value: value}},
		{"\"Rs 1,234,567.89\"", kitgo.Currency{Tag: "sd-PK", Value: value}},
		{"\"CA$ 1,234,567.89\"", kitgo.Currency{Tag: "iu-Latn-CA", Value: value}},
		{"\"DZD 1 234 567,89\"", kitgo.Currency{Tag: "tzm-Latn-DZ", Value: value}},
		{"\"₹ 1,234,567.89\"", kitgo.Currency{Tag: "ne-IN", Value: value}},
		{"\"US$ 1,234,567.89\"", kitgo.Currency{Tag: "quz-EC", Value: value}},
		{"\"Br 1,234,567.89\"", kitgo.Currency{Tag: "ti-ET", Value: value}},
		{"\"HK$ 1,234,567.89\"", kitgo.Currency{Tag: "zh-HK", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "de-AT", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-AU", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "es-ES", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1 234 567,89\"", kitgo.Currency{Tag: "fr-CA", Value: value}},
		{"\"XXX 1.234.567,89\"", kitgo.Currency{Tag: "sr-Cyrl-CS", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "se-FI", Value: value, Format: "%[1]s%[2]s"}},
		{"\"MAD 1,234,567.89\"", kitgo.Currency{Tag: "tmz-MA", Value: value}},
		{"\"PEN 1,234,567.89\"", kitgo.Currency{Tag: "quz-PE", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "zh-SG", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "de-LU", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-CA", Value: value}},
		{"\"Q 1,234,567.89\"", kitgo.Currency{Tag: "es-GT", Value: value}},
		{"\"CHF 1 234 567,89\"", kitgo.Currency{Tag: "fr-CH", Value: value}},
		{"\"KM 1.234.567,89\"", kitgo.Currency{Tag: "hr-BA", Value: value}},
		{"\"NOK 1,234,567.89\"", kitgo.Currency{Tag: "smj-NO", Value: value}},
		{"\"MOP$ 1,234,567.89\"", kitgo.Currency{Tag: "zh-MO", Value: value}},
		{"\"CHF 1’234’567.89\"", kitgo.Currency{Tag: "de-LI", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-NZ", Value: value}},
		{"\"₡ 1 234 567,89\"", kitgo.Currency{Tag: "es-CR", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "fr-LU", Value: value, Format: "%[1]s%[2]s"}},
		{"\"KM 1.234.567,89\"", kitgo.Currency{Tag: "bs-Latn-BA", Value: value}},
		{"\"SEK 1,234,567.89\"", kitgo.Currency{Tag: "smj-SE", Value: value}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "en-IE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"B/. 1,234,567.89\"", kitgo.Currency{Tag: "es-PA", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fr-MC", Value: value, Format: "%[1]s%[2]s"}},
		{"\"KM 1.234.567,89\"", kitgo.Currency{Tag: "sr-Latn-BA", Value: value}},
		{"\"NOK 1,234,567.89\"", kitgo.Currency{Tag: "sma-NO", Value: value}},
		{"\"R 1 234 567,89\"", kitgo.Currency{Tag: "en-ZA", Value: value}},
		{"\"RD$ 1,234,567.89\"", kitgo.Currency{Tag: "es-DO", Value: value}},
		//# {"\"", pkg.Currency{Tag: "fr-West", Value: value}},
		{"\"КМ 1.234.567,89\"", kitgo.Currency{Tag: "sr-Cyrl-BA", Value: value}},
		{"\"SEK 1,234,567.89\"", kitgo.Currency{Tag: "sma-SE", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-JM", Value: value}},
		{"\"Bs. 1.234.567,89\"", kitgo.Currency{Tag: "es-VE", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "fr-RE", Value: value, Format: "%[1]s%[2]s"}},
		{"\"КМ 1.234.567,89\"", kitgo.Currency{Tag: "bs-Cyrl-BA", Value: value}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "sms-FI", Value: value, Format: "%[1]s%[2]s"}},
		//# {"\"", pkg.Currency{Tag: "en-CB", Value: value}},
		{"\"$ 1.234.567,89\"", kitgo.Currency{Tag: "es-CO", Value: value}},
		{"\"FCFA 1 234 568\"", kitgo.Currency{Tag: "fr-CG", Value: value}},
		{"\"RSD 1.234.567,89\"", kitgo.Currency{Tag: "sr-Latn-RS", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "smn-FI", Value: value, Format: "%[1]s%[2]s"}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-BZ", Value: value}},
		{"\"S/ 1,234,567.89\"", kitgo.Currency{Tag: "es-PE", Value: value}},
		{"\"CFA 1 234 568\"", kitgo.Currency{Tag: "fr-SN", Value: value}},
		{"\"RSD 1.234.567,89\"", kitgo.Currency{Tag: "sr-Cyrl-RS", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "en-TT", Value: value}},
		{"\"$ 1.234.567,89\"", kitgo.Currency{Tag: "es-AR", Value: value}},
		{"\"FCFA 1 234 568\"", kitgo.Currency{Tag: "fr-CM", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "sr-Latn-ME", Value: value, Format: "%[1]s%[2]s"}},
		{"\"US$ 1,234,567.89\"", kitgo.Currency{Tag: "en-ZW", Value: value}},
		{"\"$ 1.234.567,89\"", kitgo.Currency{Tag: "es-EC", Value: value}},
		{"\"CFA 1 234 568\"", kitgo.Currency{Tag: "fr-CI", Value: value}},
		{"\"€1.234.567,89\"", kitgo.Currency{Tag: "sr-Cyrl-ME", Value: value, Format: "%[1]s%[2]s"}},
		{"\"₱ 1,234,567.89\"", kitgo.Currency{Tag: "en-PH", Value: value}},
		{"\"$ 1.234.568\"", kitgo.Currency{Tag: "es-CL", Value: value}},
		{"\"CFA 1 234 568\"", kitgo.Currency{Tag: "fr-ML", Value: value}},
		{"\"IDR 1,234,567.89\"", kitgo.Currency{Tag: "en-ID", Value: value}},
		{"\"$ 1.234.567,89\"", kitgo.Currency{Tag: "es-UY", Value: value}},
		{"\"MAD 1.234.567,89\"", kitgo.Currency{Tag: "fr-MA", Value: value}},
		{"\"HK$ 1,234,567.89\"", kitgo.Currency{Tag: "en-HK", Value: value}},
		{"\"Gs. 1.234.568\"", kitgo.Currency{Tag: "es-PY", Value: value}},
		{"\"G 1 234 567,89\"", kitgo.Currency{Tag: "fr-HT", Value: value}},
		{"\"₹  12,34,567.89\"", kitgo.Currency{Tag: "en-IN", Value: value, Format: "%[1]s  %[2]s"}},
		{"\"Bs 1.234.567,89\"", kitgo.Currency{Tag: "es-BO", Value: value}},
		{"\"$ 1,234,567.89\"", kitgo.Currency{Tag: "es-SV", Value: value}},
		{"\"L 1,234,567.89\"", kitgo.Currency{Tag: "es-HN", Value: value}},
		{"\"€1 234 567,89\"", kitgo.Currency{Tag: "smn", Value: value, Format: "%[1]s%[2]s"}},
		{"\"€1,234,567.89\"", kitgo.Currency{Tag: "sms", Value: value, Format: "%[1]s%[2]s"}},
		{"\"kr 1 234 567,89\"", kitgo.Currency{Tag: "nn", Value: value}},
		{"\"KM 1.234.567,89\"", kitgo.Currency{Tag: "bs", Value: value}},
		{"\"₼ 1.234.567,89\"", kitgo.Currency{Tag: "az-Latn", Value: value}},
		{"\"SEK 1,234,567.89\"", kitgo.Currency{Tag: "sma", Value: value}},
		{"\"сўм 1 234 567,89\"", kitgo.Currency{Tag: "uz-Cyrl", Value: value}},
		{"\"₮ 1,234,567.89\"", kitgo.Currency{Tag: "mn-Cyrl", Value: value}},
		{"\"CA$ 1,234,567.89\"", kitgo.Currency{Tag: "iu-Cans", Value: value}},
	}
	for i := range tests {
		b, err := kitgo.JSON.Marshal(tests[i].parameter)
		if string(b) != tests[i].expect {
			t.Log(tests[i].parameter.Tag)
		}
		if err != nil {
			t.Log(err.Error())
		}
		Expect(err).To(BeNil())
		Expect(string(b)).To(Equal(tests[i].expect))
	}
}

func Test_pkg_parallel(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	ctx := context.Background()

	t.Run("empty func", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
		defer cancel()

		ch := func(i int, err error) { Expect(err).NotTo(HaveOccurred()) }
		kitgo.Parallel(ctx, ch, nil, nil, nil, nil)
		kitgo.Parallel(ctx, nil, nil, nil, nil, nil)
	})
	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
		defer cancel()

		count := int64(0)
		ch := func(i int, err error) {
			if err != nil {
				_ = atomic.AddInt64(&count, 1)
				Expect(err.Error()).Should(Or(
					Equal(fmt.Errorf("error").Error()),
					Equal(context.DeadlineExceeded.Error()),
				))
			}
		}
		kitgo.Parallel(ctx, ch,
			func() error { <-time.After(10 * time.Millisecond); return nil },
			func() error { <-time.After(11 * time.Millisecond); return nil },
			func() error { <-time.After(12 * time.Millisecond); return nil },
			func() error { <-time.After(13 * time.Millisecond); return nil },
			func() error { <-time.After(14 * time.Millisecond); return nil },
			func() error { <-time.After(15 * time.Millisecond); return nil },
			func() error { return fmt.Errorf("error") },
			func() error { return fmt.Errorf("error") },
			func() error { return fmt.Errorf("error") },
			func() error { return fmt.Errorf("error") },
			func() error { return fmt.Errorf("error") },
			func() error { return fmt.Errorf("error") },
		)
		// max 7 = 6 err_ + 1 context.DeadlineExceeded
		Expect(count).To(BeNumerically("~", 1, 7))
	})
}
