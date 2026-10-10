package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerificationLink(t *testing.T) {
	const origin = "https://dev.matheo-galuba.com"

	_, err := verificationLink("https://attacker.example", origin+"/tool/ormi", "tok")
	assert.Error(t, err, "unknown origins must not receive login links")
	_, err = verificationLink("", "", "tok")
	assert.Error(t, err)

	cases := map[string]string{
		origin + "/tool/ormi?tab=stats":  "%2Ftool%2Formi%3Ftab%3Dstats",
		"https://attacker.example/phish": "%2Fphish",
		"":                               "%2F",
		"https://x//attacker.example/a":  "%2F",
		`https://x/\attacker.example`:    "%2F%255Cattacker.example",
		`"><a href="https://evil">`:      "%2F",
		"javascript:alert(1)":            "%2F",
	}
	for referer, redirect := range cases {
		link, err := verificationLink(origin, referer, "a.b-c")
		assert.NoError(t, err)
		assert.Equal(t, origin+"/verify/email?redirect="+redirect+"&token=a.b-c", link, referer)
	}
}
