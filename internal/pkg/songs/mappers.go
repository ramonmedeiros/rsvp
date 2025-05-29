package songs

import (
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

func MapSearchResultToSongs(tracks *spotify.FullTrackPage) []Song {
	if len(tracks.Tracks) == 0 {
		return nil
	}

	return lo.Map(tracks.Tracks, func(t spotify.FullTrack, _ int) Song {
		return Song{
			Artists: lo.Map(t.Artists, func(a spotify.SimpleArtist, _ int) Artist {
				return Artist{
					Name: a.Name,
				}
			}),
			Endpoint:   t.Endpoint,
			Name:       t.Name,
			Popularity: int(t.Popularity),
		}
	})
}
