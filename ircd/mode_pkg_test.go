package ircd_test

import (
	"nxircd/ircd"
	"testing"
)

func TestMode(t *testing.T) {

	t.Run("should parse modes correctly", func(t *testing.T) {
		testExamples := [][]string{
			[]string{"+o", "Testy"},
			[]string{"+s"},
			[]string{"-o", "Testy"},
			[]string{"+bb", "Dude", "Man"},
		}

		testResults := []ircd.ModeChange{
			{Action: '+', Mode: 'o', Arg: "Testy"},
			{Action: '+', Mode: 's', Arg: ""},
			{Action: '-', Mode: 'o', Arg: "Testy"},
			{Action: '+', Mode: 'b', Arg: "Dude"},
		}

		for pos, example := range testExamples {
			result := testResults[pos]

			changes := ircd.ParseCMode(example...)

			if changes[0].Action != result.Action {
				t.Fatalf("Action does not match: %s != %s", changes[0].Action, result.Action)
			}

			if changes[0].Mode != result.Mode {
				t.Fatalf("Mode does not match: %s != %s", changes[0].Mode, result.Mode)
			}

			if changes[0].Arg != result.Arg {
				t.Fatalf("Arg does not match: %s != %s", changes[0].Arg, result.Arg)
			}
		}
	})
}
