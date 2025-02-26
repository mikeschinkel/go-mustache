package gravatar_test

import (
	"testing"

	"github.com/alexkappa/mustache/gravatar"
)

func TestURLWithArgs(t *testing.T) {
	tests := []struct {
		name   string
		email  string
		rating gravatar.RatingType
		args   *gravatar.Args
		want   string
		error  bool
	}{
		{
			name:   "Basic URL",
			email:  "bozo@clownsrus.com",
			rating: "pg",
			args: &gravatar.Args{
				Width:   250,
				Default: "mp",
			},
			want: "http://gravatar.com/avatar/069578194b8b7baf0f4db0bba4499859?r=pg&s=250",
		},
		{
			name:   "Invalid Default",
			email:  "bozo@clownsrus.com",
			rating: "pg",
			args: &gravatar.Args{
				Default: "notvalid",
			},
			error: true,
			want:  "The value 'notvalid' is not a valid Gravatar default; must be one of '404', 'blank', 'color', 'default', 'identicon', 'initials', 'monsterid', 'mp', 'retro', 'robohash', 'wavatar'.  See https://docs.gravatar.com/api/avatars/images/ for more details",
		},
		{
			name:   "Invalid Rating",
			email:  "bozo@clownsrus.com",
			rating: "abc",
			args: &gravatar.Args{
				Default: "mp",
			},
			error: true,
			want:  "The value 'abc' is not a valid Gravatar rating; must be one of 'g', 'pg', 'r', 'x'.  See https://docs.gravatar.com/api/avatars/images/ for more details",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.error {
						t.Errorf("URLWithArgs() unexpected panic: %v", r)
						return
					}
					if r != tt.want {
						t.Errorf("URLWithArgs() panic = %v, want %v", r, tt.want)
					}
				}
			}()
			if gotU := gravatar.URLWithArgs(tt.email, tt.rating, tt.args); gotU != tt.want {
				t.Errorf("URLWithArgs() = %v, want %v", gotU, tt.want)
			}
		})
	}
}
