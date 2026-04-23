package service

import "testing"

func TestParsePromptFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		wantArtist string
		wantAlbum  string
	}{
		{
			name:       "chinese directives",
			query:      "我想听抒情歌，歌手:周杰伦，专辑:七里香",
			wantArtist: "周杰伦",
			wantAlbum:  "七里香",
		},
		{
			name:       "english directives",
			query:      "something upbeat artist: Coldplay album: Parachutes",
			wantArtist: "Coldplay",
			wantAlbum:  "Parachutes",
		},
		{
			name:       "directive with fullwidth colon",
			query:      "歌手：陈奕迅 专辑：认了吧",
			wantArtist: "陈奕迅",
			wantAlbum:  "认了吧",
		},
		{
			name:       "no directives",
			query:      "想听治愈系女声",
			wantArtist: "",
			wantAlbum:  "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			artist, album := parsePromptFilters(tc.query)
			if artist != tc.wantArtist {
				t.Fatalf("artist = %q; want %q", artist, tc.wantArtist)
			}
			if album != tc.wantAlbum {
				t.Fatalf("album = %q; want %q", album, tc.wantAlbum)
			}
		})
	}
}
