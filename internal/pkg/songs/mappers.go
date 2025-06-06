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
			ID: t.ID.String(),
			Artists: lo.Map(t.Artists, func(a spotify.SimpleArtist, _ int) Artist {
				return Artist{Name: a.Name}
			}),
			Endpoint:   t.Endpoint,
			Name:       t.Name,
			Popularity: int(t.Popularity),
			Images: lo.Map(t.Album.Images, func(i spotify.Image, _ int) string {
				return i.URL
			}),
		}
	})
}
