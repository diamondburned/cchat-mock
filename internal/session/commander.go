package session

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat/text"
	"github.com/pkg/errors"
)

type Commander struct{}

func (c *Commander) Run(cmds []string, w io.Writer) error {
	switch cmd := arg(cmds, 0); cmd {
	case "ls":
		fmt.Fprintln(w, "Commands: ls, random")

	case "random":
		// callback used to generate stuff and stream into readcloser
		var generator func() string
		// number of times to generate the word
		var times = 1

		switch arg(cmds, 1) {
		case "paragraph":
			generator = randomdata.Paragraph
		case "noun":
			generator = randomdata.Noun
		case "silly_name":
			generator = randomdata.SillyName
		default:
			return errors.New("Usage: random <paragraph|noun|silly_name> [repeat]")
		}

		if n := arg(cmds, 2); n != "" {
			i, err := strconv.Atoi(n)
			if err != nil {
				return errors.Wrap(err, "Failed to parse repeat number")
			}
			times = i
		}

		for i := 0; i < times; i++ {
			// Yes, we're simulating this even in something as trivial as a
			// command prompt.
			if err := internet.SimulateAustralian(); err != nil {
				fmt.Fprintln(w, "Error:", err)
				continue
			}

			fmt.Fprintln(w, generator())
		}

	default:
		return fmt.Errorf("Unknown command: %s", cmd)
	}

	return nil
}

func (s *Commander) AsCompleter() cchat.Completer { return s }

func (s *Commander) Complete(words []string, i int64) []cchat.CompletionEntry {
	switch {
	case strings.HasPrefix("ls", words[i]):
		return newCompEntries("ls")

	case strings.HasPrefix("random", words[i]):
		return newCompEntries(
			"random paragraph",
			"random noun",
			"random silly_name",
		)

	case lookbackCheck(words, i, "random", "paragraph"):
		return newCompEntries("paragraph")

	case lookbackCheck(words, i, "random", "noun"):
		return newCompEntries("noun")

	case lookbackCheck(words, i, "random", "silly_name"):
		return newCompEntries("silly_name")

	default:
		return nil
	}
}

// completion will only override `this'.
func lookbackCheck(words []string, i int64, prev, this string) bool {
	return strings.HasPrefix(this, words[i]) && i > 0 && words[i-1] == prev
}

func newCompEntries(raws ...string) []cchat.CompletionEntry {
	var entries = make([]cchat.CompletionEntry, len(raws))
	for i, raw := range raws {
		entries[i] = cchat.CompletionEntry{
			Raw:  raw,
			Text: text.Plain(raw),
		}
	}

	return entries
}

func arg(sl []string, i int) string {
	if i >= len(sl) {
		return ""
	}
	return sl[i]
}
