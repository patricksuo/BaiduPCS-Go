package converter_test

import (
	"strings"
	"testing"

	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
)

func TestTrimPathInvalidChars(t *testing.T) {
	trimmed := converter.TrimPathInvalidChars("ksjadfi*/?adf")
	if strings.Compare(trimmed, "ksjadfiadf") != 0 {
		t.Fatalf("trimmed: %s\n", trimmed)
	}
}
