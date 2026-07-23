package emailnorm

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"user@example.com", "user@example.com"},
		{"  User@Example.COM ", "user@example.com"},
		{"user+label@example.com", "user@example.com"},
		{"user+a+b@example.com", "user@example.com"},
		{"John.Doe+spam@GMail.com", "johndoe@gmail.com"},
		{"j.o.h.n@googlemail.com", "john@gmail.com"},
		{"john@googlemail.com", "john@gmail.com"},
		{"first.last@example.com", "first.last@example.com"}, // dots kept outside gmail
		{"", ""},
		{"no-at-sign", "no-at-sign"},
		{"@example.com", "@example.com"}, // empty local: passthrough
		{"user@", "user@"},               // empty domain: passthrough
		{"a@b@c.com", "a@b@c.com"},       // last @ splits: local "a@b" has no +; kept
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeCollisions(t *testing.T) {
	same := [][2]string{
		{"john.doe@gmail.com", "johndoe@gmail.com"},
		{"johndoe+wypas@gmail.com", "j.o.h.n.d.o.e@googlemail.com"},
		{"user+1@fastmail.com", "user+2@fastmail.com"},
	}
	for _, p := range same {
		if Normalize(p[0]) != Normalize(p[1]) {
			t.Errorf("expected %q and %q to normalize identically (got %q vs %q)",
				p[0], p[1], Normalize(p[0]), Normalize(p[1]))
		}
	}
	distinct := [][2]string{
		{"user@example.com", "user@example.org"},
		{"first.last@example.com", "firstlast@example.com"}, // dots significant off-gmail
	}
	for _, p := range distinct {
		if Normalize(p[0]) == Normalize(p[1]) {
			t.Errorf("expected %q and %q to stay distinct (both %q)", p[0], p[1], Normalize(p[0]))
		}
	}
}
