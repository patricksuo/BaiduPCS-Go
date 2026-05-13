package tieba

import "testing"

// 期望值由原始实现 github.com/qjfoidnh/baidu-tools@v1.2.0 (tiebautil / randominfo) 生成,
// 用于保证本次 port 与原行为完全一致。

func TestPhoneModelAndIMEI(t *testing.T) {
	if got := getPhoneModel(""); got != "LG-H818" {
		t.Errorf("getPhoneModel(\"\") = %q, want %q", got, "LG-H818")
	}
	if got := getPhoneModel("abc"); got != "G8142" {
		t.Errorf("getPhoneModel(\"abc\") = %q, want %q", got, "G8142")
	}
	if got := sumIMEI("LG-H818_"); got != 161177100287274 {
		t.Errorf("sumIMEI(\"LG-H818_\") = %d, want %d", got, uint64(161177100287274))
	}
	if got := sumIMEI("x"); got != 677458744678842 {
		t.Errorf("sumIMEI(\"x\") = %d, want %d", got, uint64(677458744678842))
	}
	if got := stringReverse("161177100287274"); got != "472782001771161" {
		t.Errorf("stringReverse = %q, want %q", got, "472782001771161")
	}
}

func TestTiebaClientSignature(t *testing.T) {
	// 与 NewUserInfoByBDUSS 中构造的 post 一致 (注意: 不含 "BDUSS" 键)
	post := map[string]string{
		"bdusstoken":  "FAKEBDUSS|null",
		"channel_id":  "",
		"channel_uid": "",
		"stErrorNums": "0",
		"subapp_type": "mini",
		"timestamp":   "1700000000922",
	}
	tiebaClientSignature(post)

	want := map[string]string{
		"_client_type":    "2",
		"_client_version": "7.0.0.0",
		"_phone_imei":     "161177100287274",
		"from":            "mini_ad_wandoujia",
		"model":           "LG-H818",
		"cuid":            "0C92F544C09528D294B3857079A22950|472782001771161",
		"sign":            "B6F33696983C6D9971190121E62F6052",
	}
	for k, v := range want {
		if post[k] != v {
			t.Errorf("post[%q] = %q, want %q", k, post[k], v)
		}
	}
}

func TestTiebaClientRawQuerySignature(t *testing.T) {
	const in = "has_plist=0&need_post_count=1&rn=1&uid=123456"
	const want = in + "&sign=CEF8B2825B695B4F54F48BF3314ED3A0"
	if got := tiebaClientRawQuerySignature(in); got != want {
		t.Errorf("tiebaClientRawQuerySignature(%q) = %q, want %q", in, got, want)
	}
}

func TestSexString(t *testing.T) {
	cases := map[int64]string{1: "♂", 2: "♀", 0: "unknown", 9: "unknown"}
	for in, want := range cases {
		if got := sexString(in); got != want {
			t.Errorf("sexString(%d) = %q, want %q", in, got, want)
		}
	}
}
