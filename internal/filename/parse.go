package filename

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	validExtension       = ".pdf"
	metadataPrefix       = "~"
	arrangersPrefix      = "_"
	keyPrefix            = "["
	keySuffix            = "]"
	withAccompaniment    = "+"
	withoutAccompaniment = "-"
)

const (
	minPrintASCII     = '\u0020'        // Minimum printable ASCII value
	maxPrintASCII     = '\u007E'        // Maximum printable ASCII value
	charactersToAvoid = "\"*./:;<>?\\|" // Characters to avoid according to cloud storage
)

// ValidateFilenameAndExtension checks that path basenames are PDF score
// filenames and returns the stem (filename without the .pdf extension).
func ValidateFilenameAndExtension(pathWithExt string) (string, error) {

	var (
		ok   bool
		stem string // Song's filename without the path but including no extension
	)

	// Only the filename is required, so any path prefixing the filename needs
	// to be removed.
	pathWithExt = filepath.Base(pathWithExt)

	// If the file doesn't end with validExtension, it is not a suitable file.
	stem, ok = strings.CutSuffix(pathWithExt, validExtension)
	if !ok {
		return "", fmt.Errorf("%q is not a PDF file", pathWithExt)
	}

	// If the filename has leading or trailing spaces, it may not be supported
	// by cloud storage.
	if stem != strings.Trim(stem, " ") {
		return "", fmt.Errorf("%q has leading or trailing spaces", stem)
	}

	// If the filename doesn't have a single instance of metadataPrefix, the
	// metadata cannot be reliably parsed.
	if strings.Count(stem, metadataPrefix) != 1 {
		return "", fmt.Errorf("%q must have a single %q in the filename to allow metadata to be parsed", pathWithExt, metadataPrefix)
	}

	// To avoid any complications with internationalisation, ensure the filename
	// is composed of 7-bit ASCII printable characters, excluding
	// charactersToAvoid.
	for i, w := 0, 0; i < len(stem); i += w {
		rune, width := utf8.DecodeRuneInString(stem[i:])
		if (rune < minPrintASCII) || (rune > maxPrintASCII) {
			return "", fmt.Errorf("%q is not a valid character in %q", rune, stem)
		}
		if strings.Contains(charactersToAvoid, string(rune)) {
			return "", fmt.Errorf("%q is not a valid character in %q", rune, stem)
		}
		w = width
	}

	return stem, nil
}

func ValidateMetadataTags(stem string) (string, error) {

	var metadata string // Song's metadata embedded in the filename

	// The metadata needs to be isolated to ensure it contains the correct
	// quantity of each tag.
	splitStem := strings.Split(stem, metadataPrefix)
	metadata = strings.Trim(splitStem[1], " ")

	// The metadata cannot be parsed if it contains the incorrect quantity of
	// each tag.
	if strings.Count(metadata, arrangersPrefix) > 1 ||
		strings.Count(metadata, keyPrefix) != 1 ||
		strings.Count(metadata, keySuffix) != 1 {
		return "", fmt.Errorf("wrong number of metadata tags")
	}

	// The tag locations are required to determine whether any
	// arrangers are included and to ensure the tags are in the correct order.
	arrangersLoc := strings.Index(metadata, arrangersPrefix)
	keyLoc := strings.Index(metadata, keyPrefix)
	keyEndLoc := strings.Index(metadata, keySuffix)

	// If arrangersPrefix is after keyPrefix, or keyPrefix is after keySuffix, the
	// metadata cannot be parsed. An absent arrangersPrefix is -1 and sorts before
	// keyPrefix.
	if (arrangersLoc > keyLoc) || (keyLoc > keyEndLoc) {
		return "", fmt.Errorf("metadata tags in wrong order")
	}

	return metadata, nil
}

func ParseComposers(metadata string, abbreviations map[string]string) (string, error) {

	var splitMetadata []string // Metadata split into slices

	// Need to isolate the Composers' metadata to use as a key for composers.
	if strings.Contains(metadata, arrangersPrefix) {
		// The Composers' metadata is before the arrangersPrefix.
		splitMetadata = strings.Split(metadata, arrangersPrefix)
	} else {
		// The Composers' metadata is before the keyPrefix.
		splitMetadata = strings.Split(metadata, keyPrefix)
	}

	return lookupAbbreviation(splitMetadata[0], abbreviations)
}

func ParseArrangers(metadata string, abbreviations map[string]string) (string, error) {

	var (
		splitMetadata []string // Metadata split into slices
	)
	// There can be no arrangers if the arrangersPrefix is not present.
	if !strings.Contains(metadata, arrangersPrefix) {
		return "", nil
	}

	// Need to isolate the Arrangers' metadata to use as a key for arrangers.
	// The Arrangers' metadata is after the arrangersPrefix ...
	splitMetadata = strings.Split(metadata, arrangersPrefix)
	// ... and before the keyPrefix
	splitMetadata = strings.Split(splitMetadata[1], keyPrefix)

	return lookupAbbreviation(splitMetadata[0], abbreviations)
}

func lookupAbbreviation(abbreviation string, abbreviations map[string]string) (string, error) {

	var (
		ok   bool
		name string // Name(s) of composer(s) or arranger(s)
	)

	abbreviation = strings.Trim(abbreviation, " ")
	if len(abbreviation) == 0 {
		return "", nil
	}

	name, ok = abbreviations[abbreviation]
	if !ok {
		return "", fmt.Errorf("%q not found", abbreviation)
	}

	return strings.TrimSpace(name), nil
}

func ParseKey(metadata string) (string, error) {

	var (
		ok            bool
		splitMetadata []string // Metadata split into slices
		key           string   // Song's initial key signature
	)

	keys := map[string]string{
		"":    "", // No key specified, for example for an atonal score.
		"Cb":  "keysf:-7, keymi:0",
		"Abm": "keysf:-7, keymi:1",
		"Gb":  "keysf:-6, keymi:0",
		"Ebm": "keysf:-6, keymi:1",
		"Db":  "keysf:-5, keymi:0",
		"Bbm": "keysf:-5, keymi:1",
		"Ab":  "keysf:-4, keymi:0",
		"Fm":  "keysf:-4, keymi:1",
		"Eb":  "keysf:-3, keymi:0",
		"Cm":  "keysf:-3, keymi:1",
		"Bb":  "keysf:-2, keymi:0",
		"Gm":  "keysf:-2, keymi:1",
		"F":   "keysf:-1, keymi:0",
		"Dm":  "keysf:-1, keymi:1",
		"C":   "keysf:0, keymi:0",
		"Am":  "keysf:0, keymi:1",
		"G":   "keysf:1, keymi:0",
		"Em":  "keysf:1, keymi:1",
		"D":   "keysf:2, keymi:0",
		"Bm":  "keysf:2, keymi:1",
		"A":   "keysf:3, keymi:0",
		"F#m": "keysf:3, keymi:1",
		"E":   "keysf:4, keymi:0",
		"C#m": "keysf:4, keymi:1",
		"B":   "keysf:5, keymi:0",
		"G#m": "keysf:5, keymi:1",
		"F#":  "keysf:6, keymi:0",
		"D#m": "keysf:6, keymi:1",
		"C#":  "keysf:7, keymi:0",
		"A#m": "keysf:7, keymi:1",
		"7#":  "",
		"6#":  "",
		"5#":  "",
		"4#":  "",
		"3#":  "",
		"2#":  "",
		"1#":  "",
		"0":   "",
		"1b":  "",
		"2b":  "",
		"3b":  "",
		"4b":  "",
		"5b":  "",
		"6b":  "",
		"7b":  "",
	}

	// Need to isolate the Initial Key Signature's metadata to use as the key
	// for keys. The Initial Key Signature's metadata is after the keyPrefix ...
	splitMetadata = strings.Split(metadata, keyPrefix)
	// ... and before the keySuffix.
	splitMetadata = strings.Split(splitMetadata[1], keySuffix)
	raw := strings.Trim(splitMetadata[0], " ")
	key, ok = keys[raw]
	if !ok {
		return "", fmt.Errorf("%q is not a valid key signature", raw)
	}

	return key, nil
}

func ParseAccompaniment(metadata string) (string, error) {

	var (
		splitMetadata []string // Metadata split into slices
		accMetadata   string   // Metadata indicating whether or not the song has an accompaniment
	)

	// The Accompaniment's metadata is a single character after the keySuffix.
	splitMetadata = strings.Split(metadata, keySuffix)
	accMetadata = strings.Trim(splitMetadata[1], " ")
	switch accMetadata {
	case withAccompaniment:
		return "With Accompaniment", nil
	case withoutAccompaniment:
		return "Without Accompaniment", nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("%q is not valid for accompaniment", accMetadata)
	}
}
