// Copyright (c) 2025 Mike Schinkel

package gravatar

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

const (
	HelpURL      = "https://docs.gravatar.com/api/avatars/images/"
	URLFormat    = "http://gravatar.com/avatar/%x?r=%s&s=%d"
	DefaultWidth = 200
	MinWidth     = 1
	MaxWidth     = 2048
)

type ImageType string

const (
	DefaultImageType   = `default`   // uses the Gravatar logo
	InitialsImageType  = `initials`  // uses the profile name as initials, with a generated background and foreground color (beta). There are additional parameters, described later.
	ColorImageType     = `color`     // a generated color (beta)
	NotFoundImageType  = `404`       // do not load any image if none is associated with the email hash, instead return an HTTP 404 (File Not Found) response
	MpImageType        = `mp`        // (mystery-person) a simple, cartoon-style silhouetted outline of a person (does not vary by email hash)
	IdenticonImageType = `identicon` // a geometric pattern based on an email hash
	MonsterIdImageType = `monsterid` // a generated ‘monster’ with different colors, faces, etc
	WavatarImageType   = `wavatar`   // generated faces with differing features and backgrounds
	RetroImageType     = `retro`     // awesome generated, 8-bit arcade-style pixelated faces
	RobohashImageType  = `robohash`  // a generated robot with different colors, faces, etc
	BlankImageType     = `blank`     // a transparent PNG image (border added to HTML below for demonstration purposes)
)

var AllImageTypesMap = map[ImageType]struct{}{
	DefaultImageType:   {},
	InitialsImageType:  {},
	ColorImageType:     {},
	NotFoundImageType:  {},
	MpImageType:        {},
	IdenticonImageType: {},
	MonsterIdImageType: {},
	WavatarImageType:   {},
	RetroImageType:     {},
	RobohashImageType:  {},
	BlankImageType:     {},
}

type RatingType string

const (
	GRatingType  = `g`  // suitable for display on all websites with any audience type.
	PGRatingType = `pg` // may contain rude gestures, provocatively dressed ls, the lesser swear words, or mild violence.
	RRatingType  = `r`  // may contain such things as harsh profanity, intense violence, nudity, or hard drug use.
	XRatingType  = `x`  // may contain sexual imagery or extremely disturbing violence.
)

var AllRatingTypeMap = map[RatingType]struct{}{
	GRatingType:  {},
	PGRatingType: {},
	RRatingType:  {},
	XRatingType:  {},
}

type Args struct {
	Width   int
	Default ImageType
}

//goland:noinspection GoUnusedExportedFunction
func URL(email string, rating RatingType) (u string) {
	return URLWithArgs(email, rating, nil)
}

func URLWithArgs(email string, rating RatingType, args *Args) (u string) {
	if rating != "" {
		rating = RatingType(strings.ToLower(string(rating)))
	}
	if _, ok := AllRatingTypeMap[rating]; !ok {
		oops("rating", rating, toStrings(AllRatingTypeMap))
	}
	if args.Default == "" {
		args.Default = DefaultImageType
	}
	if _, ok := AllImageTypesMap[args.Default]; !ok {
		oops("default", args.Default, toStrings(AllImageTypesMap))
	}
	if args.Width == 0 {
		args.Width = DefaultWidth
	}
	if args.Width < 1 || args.Width > 2048 {
		panic(fmt.Sprintf("The value '%d' is not a valid Gravatar with; must be betweeb %d and %d, inclusive",
			args.Width,
			MinWidth,
			MaxWidth,
		))
	}
	return fmt.Sprintf(URLFormat, md5.Sum([]byte(email)), rating, args.Width)
}

func toStrings[T ~string](m map[T]struct{}) []string {
	ss := make([]string, 0, len(m))
	for k := range m {
		ss = append(ss, string(k))
	}
	sort.Strings(ss)
	return ss
}
func oops(name string, value any, options []string) {
	panic(fmt.Sprintf("The value '%v' is not a valid Gravatar %s; must be one of '%s'.  See %s for more details",
		value,
		name,
		strings.Join(options, `', '`),
		HelpURL,
	))
}
