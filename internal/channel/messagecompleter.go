package channel

import (
	"strings"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/message"
)

type MessageCompleter struct {
	msgr *Messenger
}

func (msgc MessageCompleter) Complete(words []string, i int64) []cchat.CompletionEntry {
	switch {
	case strings.HasPrefix("complete", words[i]):
		return makeCompletion(
			"complete",
			"complete me",
			"complete you",
			"complete everyone",
		)

	case lookbackCheck(words, i, "complete", "me"):
		return makeCompletion("me")

	case lookbackCheck(words, i, "complete", "you"):
		return makeCompletion("you")

	case lookbackCheck(words, i, "complete", "everyone"):
		return makeCompletion("everyone")

	case lookbackCheck(words, i, "best", "femboys:"):
		return makeCompletion(
			"trap: Astolfo",
			"trap: Hackadoll No. 3",
			"trap: Totsuka",
			"trap: Felix Argyle",
		)

	default:
		var found = map[string]struct{}{}

		msgc.msgr.messageMutex.Lock()
		defer msgc.msgr.messageMutex.Unlock()

		var entries []cchat.CompletionEntry

		// Look for members.
		for _, id := range msgc.msgr.messageids {
			if msg := msgc.msgr.messages[id]; strings.HasPrefix(msg.AuthorName(), words[i]) {
				if _, ok := found[msg.AuthorName()]; ok {
					continue
				}

				found[msg.AuthorName()] = struct{}{}

				entries = append(entries, cchat.CompletionEntry{
					Raw:     msg.AuthorName(),
					Text:    msg.Author().Name(),
					IconURL: msg.Author().Avatar(),
				})
			}
		}

		return entries
	}

	return nil
}

func makeCompletion(word ...string) []cchat.CompletionEntry {
	var entries = make([]cchat.CompletionEntry, len(word))
	for i, w := range word {
		entries[i].Raw = w
		entries[i].Text.Content = w
		entries[i].IconURL = message.AvatarURL
	}
	return entries
}

// completion will only override `this'.
func lookbackCheck(words []string, i int64, prev, this string) bool {
	return strings.HasPrefix(this, words[i]) && i > 0 && words[i-1] == prev
}
