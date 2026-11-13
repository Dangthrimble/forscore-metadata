package namesedit

import "github.com/Dangthrimble/forscore-metadata/internal/names"

// Kind selects which abbreviation dictionary the local editor maintains.
type Kind struct {
	Command      string
	Usage        string
	PageTitle    string
	Heading      string
	ListHeading  string
	ValueHeading string
	ValueLabel   string
	SearchHint   string
	CommaHint    string
	ChooseTitle  string
	CreateTitle  string
	DefaultFile  string
	Resolve      func(string) (string, error)
	Configured   func(*names.Config) string
	SetPath      func(string) error
}

// Names edits the composer and arranger dictionary.
var Names = Kind{
	Command:      "names",
	Usage:        "fn2fm names [path]",
	PageTitle:    "fn2fm names editor",
	Heading:      "fn2fm names",
	ListHeading:  "Names",
	ValueHeading: "Name",
	ValueLabel:   "Full name(s)",
	SearchHint:   "Name or abbreviation",
	CommaHint:    "Separate people with commas, for example Chris Tomlin, Ed Cash, Matt Redman.",
	ChooseTitle:  "Choose names.json",
	CreateTitle:  "Create names.json",
	DefaultFile:  "names.json",
	Resolve:      names.ResolvePath,
	Configured: func(c *names.Config) string {
		if c == nil {
			return ""
		}
		return c.NamesPath
	},
	SetPath: names.SetNamesPath,
}

// Genres edits the genre dictionary used when PDF Subject is genres.
var Genres = Kind{
	Command:      "genres",
	Usage:        "fn2fm genres [path]",
	PageTitle:    "fn2fm genres editor",
	Heading:      "fn2fm genres",
	ListHeading:  "Genres",
	ValueHeading: "Genre",
	ValueLabel:   "Genre(s)",
	SearchHint:   "Genre or abbreviation",
	CommaHint:    "Separate genres with commas, for example Christmas, Folk.",
	ChooseTitle:  "Choose genres.json",
	CreateTitle:  "Create genres.json",
	DefaultFile:  "genres.json",
	Resolve:      names.ResolveGenresPath,
	Configured: func(c *names.Config) string {
		if c == nil {
			return ""
		}
		return c.GenresPath
	},
	SetPath: names.SetGenresPath,
}
