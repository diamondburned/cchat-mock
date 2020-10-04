package message

import (
	"github.com/diamondburned/aqs"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/segments"
	"github.com/diamondburned/cchat/text"
)

const AvatarURL = "" +
	"https://gist.github.com/diamondburned/" +
	"945744c2b5ce0aa0581c9267a4e5cf24/raw/" +
	"598069da673093aaca4cd4aa0ede1a0e324e9a3a/" +
	"astolfo_selfie.png"

type Author struct {
	name text.Rich
	char aqs.Character
}

var _ cchat.Author = (*Author)(nil)

func NewAuthor(name text.Rich) Author {
	return Author{name: name}
}

func RandomAuthor() Author {
	var char = aqs.RandomCharacter()
	return Author{
		char: char,
		name: text.Rich{
			Content: char.Name,
			Segments: []text.Segment{
				segments.NewColorfulSegment(char.Name, char.NameColor()),
			},
		},
	}
}

func (a Author) ID() string {
	return a.name.Content
}

func (a Author) Name() text.Rich {
	return a.name
}

func (a Author) Avatar() string {
	if a.char.ImageURL != "" {
		return a.char.ImageURL
	}
	return AvatarURL
}

// Equal returns true if this author is the same as the given other author.
func (a Author) Equal(other Author) bool {
	return a.name.Content == other.name.Content
}
