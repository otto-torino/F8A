package utils

import (
	"testing"

	"github.com/spf13/viper"
)

func TestBuildShellCommand(t *testing.T) {
	cases := []struct {
		name    string
		init    string
		command string
		want    string
	}{
		{"no init", "", "yarn build", "yarn build"},
		{"whitespace init", "   ", "yarn build", "yarn build"},
		{"with init", ". $HOME/.nvm/nvm.sh", "yarn build", ". $HOME/.nvm/nvm.sh && yarn build"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			viper.Set("shell.init", c.init)
			defer viper.Set("shell.init", "")
			if got := buildShellCommand(c.command); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
