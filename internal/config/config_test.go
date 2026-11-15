package config

import (
	"strings"
	"testing"
)

func TestCollectKeys(t *testing.T) {
	env := []string{
		"PATH=/bin",
		"SUPERCODE_API_KEY3=c",
		"SUPERCODE_API_KEY1=a",
		"SUPERCODE_API_KEY10=j",
		"SUPERCODE_API_KEY2= b ",
		"SUPERCODE_API_KEY4=a",
		"SUPERCODE_API_KEY5=",
		"SUPERCODE_API_KEYX=nope",
		"SUPERCODE_API_KEY=legacy",
	}

	got := strings.Join(collectKeys(env), ",")
	if want := "legacy,a,b,c,j"; got != want {
		t.Errorf("collectKeys = %s, want %s", got, want)
	}
}
