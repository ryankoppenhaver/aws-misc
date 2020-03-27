package main

import (
	"regexp"
	"testing"
)

// ssh-keygen -t ed25519 -f /tmp/testca -C 'test CA' -N ''
const caKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBuIgZ1YyuR5sDkrfFyByAi2FNr4e9Rwc/4Ta+ZZLzCCwAAAJCf3tWRn97V
kQAAAAtzc2gtZWQyNTUxOQAAACBuIgZ1YyuR5sDkrfFyByAi2FNr4e9Rwc/4Ta+ZZLzCCw
AAAEAMuEUZr5iOyB6C0g2J99+6MeJUYEN13Jvhdg7RTTSySm4iBnVjK5HmwOSt8XIHICLY
U2vh71HBz/hNr5lkvMILAAAAB3Rlc3QgQ0EBAgMEBQY=
-----END OPENSSH PRIVATE KEY-----`

func TestSigningSuccess(t *testing.T) {
	// ssh-keygen -t ed25519 -f /tmp/testkey -N ''
	pubKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDyJwzFhmVkHOt/dX7tbGzcL5kr8WbKCrCnF2y5B6C81 root@host"

	s, err := NewSigner(caKey)
	if err != nil {
		t.Fatal(err)
	}

	res, err := s.MakeCertificate(pubKey, []string{"i-1234.example.com", "test.example.com"})
	if err != nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile("ssh-ed25519-cert-v01@openssh.com [a-zA-Z0-9/+=]+ i-1234.example.com at \\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z").MatchString(res) {
		t.Fatalf("output %q does not match expected pattern", res)
	}

	//TODO: actually check sig
}
