package projectverify

import "testing"

// Verifies the GUID samples taken from a real Tuanjie 1.9.1 project, and hex and broken values
// around them, are accepted or rejected as Unity and Tuanjie do, with the expected duplicate key.
func TestClassifyGUIDSamples(t *testing.T) {
	samples := []struct {
		value      string
		valid      bool
		normalized string
	}{
		{value: "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M=", valid: true},
		{value: "DylK5i+tBy3hwBpolH4EHssf3fnR6QNZrkcrOW/pENa5cIxYicJoEC4=", valid: true},
		{value: "B30dsC+rVX2QA/jILTXpzr5jSTOpI7sSbFnlEF4QoQ2qB9dfRF5p1VQ=", valid: true},
		{value: "Cn0Wsi2uUns4ZwUV2RC4cE6k3wb8St45+jCfFtASi2ULX30At9b1/9k=", valid: true},
		{value: "B38Xti/4AH81anwEOACN4MNEd5K8Wr+20yEbSxVHbaYO5TiFiswlT2s=", valid: true},
		{value: "DXof4HisBnPkJWlP32wYq0aft5L5kx282zcHImRk5ZPDoZLD7PH+cA0=", valid: true},
		{value: "B3IbsyP8BinDeX7IK++D0O5jq0aDutI+yasGXGEm9m5agALoVL7ZhWs=", valid: true},
		{value: "DikbvCmsUioPnMdHYLVyexmoateqwyYuFP58/7oNV06/caMsIdjhuME=", valid: true},
		{value: "D3Mb5in7ASkpgG8PjCDSBbU4l3iFUHGQubVCKno3/xvdxGUXr9QKGIo=", valid: true},
		{value: "WX8esSypUy1Q0aysVyjJYp3h2U6d8EMHNMMEevti59VKtuvaE+phepI=", valid: true},
		{value: "BiwYsir5ACge/MGgu5y5pA8hhBa6FZUmAARe7c5XSCFksyvXJEfYDls=", valid: true},
		{value: "XHMa4SytAi5TjhUxob6NYy7Y9g971Yfvdgsahyg4IGYf6lt8EUFIOc4=", valid: true},
		{value: "XH4b4yP5VHn+SIVVouMA9b102P0k7o5jF6wJKN/h3O6O/sqo5kM+PPs=", valid: true},
		{value: "DXMf5i6sAn4lenZMigpBh1o5VO/jB3tPmkyjo8oPCuo/Omwb/qpI5MM=", valid: true},
		{value: "WX1LvX+oV3/HEGIix6XntmtDssoOy0wPhp+rDVQ+lWhlO2lcEjkMRLM=", valid: true},
		{value: "Bn8csC6oVy+iejS/GBoG32OVS+eWj1UZxhjvE+kqC0ZJ7/0RsyiEN0Y=", valid: true},
		{value: "0123456789abcdef0123456789abcdef", valid: true, normalized: "0123456789abcdef0123456789abcdef"},
		{value: "ABCDEF0123456789ABCDEF0123456789", valid: true, normalized: "abcdef0123456789abcdef0123456789"},
		{value: "00000000000000000000000000000000"},
		{value: ""},
		{value: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
		{value: "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M"},
		{value: "!nIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M="},
	}
	for _, sample := range samples {
		want := sample.normalized
		if sample.valid && want == "" {
			want = sample.value
		}
		key, form := classifyGUID(sample.value)
		valid := form == guidFormHex || form == guidFormTuanjieEncoded
		if valid != sample.valid {
			t.Errorf("classifyGUID(%q) form = %d, want valid %v", sample.value, form, sample.valid)
			continue
		}
		if valid && key != want {
			t.Errorf("classifyGUID(%q) key = %q, want %q", sample.value, key, want)
		}
	}
}

// Verifies values with the encoded shape but a prefix Tuanjie never writes get their own form, while
// values without that shape do not.
func TestClassifyGUIDSeparatesEncodedShapeFromOtherInvalidValues(t *testing.T) {
	cases := []struct {
		value string
		want  guidForm
	}{
		{value: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", want: guidFormTuanjieEncodedMismatch},
		{value: "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M", want: guidFormInvalid},
		{value: "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M==", want: guidFormInvalid},
		{value: "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv=M=", want: guidFormInvalid},
		{value: "BnIbtiz-Ui_6fZL__VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M=", want: guidFormInvalid},
	}
	for _, c := range cases {
		if _, form := classifyGUID(c.value); form != c.want {
			t.Errorf("classifyGUID(%q) form = %d, want %d", c.value, form, c.want)
		}
	}
}
