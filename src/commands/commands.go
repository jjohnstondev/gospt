package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gospt/src/gctx"

	"github.com/zmb3/spotify/v2"
)

func SetVolume(ctx *gctx.Context, client *spotify.Client, vol int) error {
	return client.Volume(ctx, vol)
}

func SetPosition(ctx *gctx.Context, client *spotify.Client, pos int) error {
	err := client.Seek(ctx, pos)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func Seek(ctx *gctx.Context, client *spotify.Client, fwd bool) error {
	current, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	newPos := current.Progress + 5000
	if !fwd {
		newPos = current.Progress - 5000
	}
	err = client.Seek(ctx, newPos)
	if err != nil {
		return err
	}
	return nil
}

func ChangeVolume(ctx *gctx.Context, client *spotify.Client, vol int) error {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return err
	}
	newVolume := state.Device.Volume + vol
	if newVolume > 100 {
		newVolume = 100
	}
	if newVolume < 0 {
		newVolume = 0
	}
	return client.Volume(ctx, newVolume)
}

func Play(ctx *gctx.Context, client *spotify.Client) error {
	err := client.Play(ctx)
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.PlayOpt(ctx, &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func ActiveDeviceExists(ctx *gctx.Context, client *spotify.Client) bool {
	current, err := client.PlayerDevices(ctx)
	if err != nil {
		return false
	}
	for _, dev := range current {
		if dev.Active {
			return true
		}
	}
	return false
}

func UserArtists(ctx *gctx.Context, client *spotify.Client, page int) (*spotify.FullArtistCursorPage, error) {
	artists, err := client.CurrentUsersFollowedArtists(ctx, spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	return artists, nil
}

func ArtistAlbums(ctx *gctx.Context, client *spotify.Client, artist spotify.ID, page int) (*spotify.SimpleAlbumPage, error) {
	albums, err := client.GetArtistAlbums(ctx, artist, []spotify.AlbumType{1, 2, 3, 4}, spotify.Market(spotify.CountryUSA), spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func Search(ctx *gctx.Context, client *spotify.Client, search string, page int) (*spotify.SearchResult, error) {
	result, err := client.Search(ctx, search, spotify.SearchTypeAlbum|spotify.SearchTypeArtist|spotify.SearchTypeTrack|spotify.SearchTypePlaylist, spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func AlbumTracks(ctx *gctx.Context, client *spotify.Client, album spotify.ID, page int) (*spotify.SimpleTrackPage, error) {
	tracks, err := client.GetAlbumTracks(ctx, album, spotify.Limit(50), spotify.Offset((page-1)*50), spotify.Market(spotify.CountryUSA))
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

func UserAlbums(ctx *gctx.Context, client *spotify.Client, page int) (*spotify.SavedAlbumPage, error) {
	albums, err := client.CurrentUsersAlbums(ctx, spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func PlayUrl(ctx *gctx.Context, client *spotify.Client, args []string) error {
	url, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	track_id := strings.Split(url.Path, "/")[2]
	err = client.QueueSong(ctx, spotify.ID(track_id))
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.QueueSongOpt(ctx, spotify.ID(track_id), &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
			err = client.NextOpt(ctx, &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	err = client.Next(ctx)
	if err != nil {
		return err
	}
	return nil
}

func QueueSong(ctx *gctx.Context, client *spotify.Client, id spotify.ID) error {
	err := client.QueueSong(ctx, id)
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.QueueSongOpt(ctx, id, &spotify.PlayOptions{
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	return nil
}

func PlaySongInPlaylist(ctx *gctx.Context, client *spotify.Client, context *spotify.URI, offset int) error {
	e := client.PlayOpt(ctx, &spotify.PlayOptions{
		PlaybackOffset:  &spotify.PlaybackOffset{Position: offset},
		PlaybackContext: context,
	})
	if e != nil {
		if isNoActiveError(e) {
			deviceID, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.PlayOpt(ctx, &spotify.PlayOptions{
				PlaybackOffset:  &spotify.PlaybackOffset{Position: offset},
				PlaybackContext: context,
				DeviceID:        &deviceID,
			})

			if err != nil {
				if isNoActiveError(err) {
					deviceID, err := activateDevice(ctx, client)
					if err != nil {
						return err
					}
					err = client.PlayOpt(ctx, &spotify.PlayOptions{
						PlaybackOffset:  &spotify.PlaybackOffset{Position: offset},
						PlaybackContext: context,
						DeviceID:        &deviceID,
					})
					if err != nil {
						return err
					}
				}
			}
			err = client.Play(ctx)
			if err != nil {
				return err
			}
		} else {
			return e
		}
	}
	return nil
}

func PlayLikedSongs(ctx *gctx.Context, client *spotify.Client, position int) error {
	err := ClearRadio(ctx, client)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	playlist, err := GetRadioPlaylist(ctx, client, "Saved Songs")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	songs, err := client.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset(position))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	to_add := []spotify.ID{}
	for _, song := range songs.Tracks {
		to_add = append(to_add, song.ID)
	}
	_, err = client.AddTracksToPlaylist(ctx, playlist.ID, to_add...)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = client.PlayOpt(ctx, &spotify.PlayOptions{
		PlaybackContext: &playlist.URI,
		PlaybackOffset: &spotify.PlaybackOffset{
			Position: 0,
		},
	})
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.PlayOpt(ctx, &spotify.PlayOptions{
				PlaybackContext: &playlist.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: 0,
				},
				DeviceID: &deviceID,
			})
			if err != nil {
				return err
			}
		}
	}
	for page := 2; page <= 5; page++ {
		songs, err := client.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset((50*(page-1))+position))
		if err != nil {
			return err
		}
		to_add := []spotify.ID{}
		for _, song := range songs.Tracks {
			to_add = append(to_add, song.ID)
		}
		client.AddTracksToPlaylist(ctx, playlist.ID, to_add...)
	}

	return err
}

func RadioGivenArtist(ctx *gctx.Context, client *spotify.Client, artist spotify.SimpleArtist) error {
	seed := spotify.Seeds{
		Artists: []spotify.ID{artist.ID},
	}
	recomendations, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	err = ClearRadio(ctx, client)
	if err != nil {
		return err
	}
	radioPlaylist, err := GetRadioPlaylist(ctx, client, artist.Name)
	if err != nil {
		return err
	}
	queue := []spotify.ID{}
	all_recs := map[spotify.ID]bool{}
	for _, rec := range recomendationIds {
		all_recs[rec] = true
		queue = append(queue, rec)
	}
	_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	client.PlayOpt(ctx, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
		PlaybackOffset: &spotify.PlaybackOffset{
			Position: 0,
		},
	})
	err = client.Repeat(ctx, "context")
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additional_recs, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additional_recs.Tracks {
			additionalRecsIds = append(additionalRecsIds, song.ID)
		}
		_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}

func RadioGivenSong(ctx *gctx.Context, client *spotify.Client, song spotify.SimpleTrack, pos int) error {
	start := time.Now().UnixMilli()
	seed := spotify.Seeds{
		Tracks: []spotify.ID{song.ID},
	}
	recomendations, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(99))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	err = ClearRadio(ctx, client)
	if err != nil {
		return err
	}
	radioPlaylist, err := GetRadioPlaylist(ctx, client, song.Name)
	if err != nil {
		return err
	}
	queue := []spotify.ID{song.ID}
	all_recs := map[spotify.ID]bool{}
	all_recs[song.ID] = true
	for _, rec := range recomendationIds {
		all_recs[rec] = true
		queue = append(queue, rec)
	}
	_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	delay := time.Now().UnixMilli() - start
	if pos != 0 {
		pos = pos + int(delay)
	}
	err = client.PlayOpt(ctx, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
		PlaybackOffset: &spotify.PlaybackOffset{
			Position: 0,
		},
		PositionMs: pos,
	})
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.PlayOpt(ctx, &spotify.PlayOptions{
				PlaybackContext: &radioPlaylist.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: 0,
				},
				DeviceID:   &deviceID,
				PositionMs: pos,
			})
			if err != nil {
				return err
			}
		}
	}
	err = client.Repeat(ctx, "context")
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additional_recs, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additional_recs.Tracks {
			additionalRecsIds = append(additionalRecsIds, song.ID)
		}
		_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}

func Radio(ctx *gctx.Context, client *spotify.Client) error {
	rand.Seed(time.Now().Unix())
	current_song, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	var seed_song spotify.SimpleTrack

	if current_song.Item != nil {
		seed_song = current_song.Item.SimpleTrack
	}
	if current_song.Item == nil {
		_, err := activateDevice(ctx, client)
		if err != nil {
			return err
		}
		tracks, err := client.CurrentUsersTracks(ctx, spotify.Limit(10))
		if err != nil {
			return err
		}
		seed_song = tracks.Tracks[rand.Intn(len(tracks.Tracks))].SimpleTrack
	} else {
		if !current_song.Playing {

			tracks, err := client.CurrentUsersTracks(ctx, spotify.Limit(10))
			if err != nil {
				return err
			}
			seed_song = tracks.Tracks[rand.Intn(len(tracks.Tracks))].SimpleTrack
		}
	}
	return RadioGivenSong(ctx, client, seed_song, current_song.Progress)
}

func RefillRadio(ctx *gctx.Context, client *spotify.Client) error {
	status, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	if !status.Playing {
		return nil
	}
	to_remove := []spotify.ID{}
	radioPlaylist, err := GetRadioPlaylist(ctx, client, "")
	if status.PlaybackContext.URI != radioPlaylist.URI {
		return nil
	}
	playlistItems, err := client.GetPlaylistItems(ctx, radioPlaylist.ID)
	if err != nil {
		return err
	}
	found := false
	page := 0
	for !found {
		tracks, err := client.GetPlaylistItems(ctx, radioPlaylist.ID, spotify.Limit(50), spotify.Offset(page*50))
		if err != nil {
			return err
		}
		for _, track := range tracks.Items {
			if track.Track.Track.ID == status.Item.ID {
				found = true
				break
			}
			to_remove = append(to_remove, track.Track.Track.ID)
		}
		page++
	}
	if len(to_remove) > 0 {
		var trackGroups []spotify.ID
		for idx, item := range to_remove {
			if idx%100 == 0 {
				_, err = client.RemoveTracksFromPlaylist(ctx, radioPlaylist.ID, trackGroups...)
				trackGroups = []spotify.ID{}
			}
			trackGroups = append(trackGroups, item)
		}
		_, err = client.RemoveTracksFromPlaylist(ctx, radioPlaylist.ID, trackGroups...)
	}
	to_add := 500 - (playlistItems.Total - len(to_remove))
	rand.Seed(time.Now().Unix())
	playlistItems, err = client.GetPlaylistItems(ctx, radioPlaylist.ID)
	if err != nil {
		return err
	}
	total := playlistItems.Total
	pages := int(math.Ceil(float64(total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(int(pages-1)) + 1
	}
	playlistPage, err := client.GetPlaylistItems(ctx, radioPlaylist.ID, spotify.Limit(50), spotify.Offset((randomPage-1)*50))
	if err != nil {
		return err
	}
	pageSongs := playlistPage.Items
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 5
	if len(pageSongs) < seedCount {
		seedCount = len(pageSongs)
	}
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.Track.Track.ID)
	}
	seed := spotify.Seeds{
		Tracks: seedIds,
	}
	recomendations, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(95))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	queue := []spotify.ID{}
	queue = append(queue, seedIds...)
	all_recs := map[spotify.ID]struct{}{}
	for _, id := range seedIds {
		all_recs[id] = struct{}{}
	}
	for idx, rec := range recomendationIds {
		if idx > to_add {
			break
		}
		all_recs[rec] = struct{}{}
		queue = append(queue, rec)
	}
	to_add -= len(queue)
	_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	err = client.Repeat(ctx, "context")
	if err != nil {
		return err
	}
	for to_add > 0 {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additional_recs, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for idx, song := range additional_recs.Tracks {
			if _, ok := all_recs[song.ID]; !ok {
				if idx > to_add {
					break
				}
				all_recs[song.ID] = struct{}{}
				additionalRecsIds = append(additionalRecsIds, song.ID)
				queue = append(queue, song.ID)
			}
		}
		to_add -= len(queue)
		_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}

func ClearRadio(ctx *gctx.Context, client *spotify.Client) error {
	radioPlaylist, err := GetRadioPlaylist(ctx, client, "")
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = client.UnfollowPlaylist(ctx, radioPlaylist.ID)
	if err != nil {
		return err
	}
	configDir, _ := os.UserConfigDir()
	os.Remove(filepath.Join(configDir, "gospt/radio.json"))
	client.Pause(ctx)
	return nil
}

func Devices(ctx *gctx.Context, client *spotify.Client) error {
	devices, err := client.PlayerDevices(ctx)
	if err != nil {
		return err
	}
	return PrintDevices(devices)
}

func Pause(ctx *gctx.Context, client *spotify.Client) error {
	err := client.Pause(ctx)
	if err != nil {
		return err
	}
	return nil
}

func TogglePlay(ctx *gctx.Context, client *spotify.Client) error {
	current, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	if !current.Playing {
		return Play(ctx, client)
	}
	return Pause(ctx, client)
}

func Like(ctx *gctx.Context, client *spotify.Client) error {
	playing, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	err = client.AddTracksToLibrary(ctx, playing.Item.ID)
	if err != nil {
		return err
	}
	return nil
}

func Unlike(ctx *gctx.Context, client *spotify.Client) error {
	playing, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	err = client.RemoveTracksFromLibrary(ctx, playing.Item.ID)
	if err != nil {
		return err
	}
	return nil
}

func Next(ctx *gctx.Context, client *spotify.Client, amt int) error {
	if amt == 1 {
		err := client.Next(ctx)
		if err != nil {
			if isNoActiveError(err) {
				deviceId, err := activateDevice(ctx, client)
				if err != nil {
					return err
				}
				err = client.NextOpt(ctx, &spotify.PlayOptions{
					DeviceID: &deviceId,
				})
				if err != nil {
					return err
				}
			}
			return err
		}
		return nil
	}
	// found := false
	// playingIndex := 0
	current, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	playbackContext := current.PlaybackContext.Type
	switch playbackContext {
	case "playlist":
		found := false
		currentTrackIndex := 0
		for !found {
			page := 1
			playlist, err := client.GetPlaylistItems(ctx, spotify.ID(strings.Split(string(current.PlaybackContext.URI), ":")[2]), spotify.Limit(50), spotify.Offset((page-1)*50))
			if err != nil {
				return err
			}
			for idx, track := range playlist.Items {
				if track.Track.Track.ID == current.Item.ID {
					currentTrackIndex = idx + (50 * (page - 1))
					found = true
					break
				}
			}
			page++
		}
		client.PlayOpt(ctx, &spotify.PlayOptions{
			PlaybackContext: &current.PlaybackContext.URI,
			PlaybackOffset: &spotify.PlaybackOffset{
				Position: currentTrackIndex + amt,
			},
		})
		return nil
	case "album":
		found := false
		currentTrackIndex := 0
		for !found {
			page := 1
			playlist, err := client.GetAlbumTracks(ctx, spotify.ID(strings.Split(string(current.PlaybackContext.URI), ":")[2]), spotify.Limit(50), spotify.Offset((page-1)*50))
			if err != nil {
				return err
			}
			for idx, track := range playlist.Tracks {
				if track.ID == current.Item.ID {
					currentTrackIndex = idx + (50 * (page - 1))
					found = true
					break
				}
			}
			page++
		}
		client.PlayOpt(ctx, &spotify.PlayOptions{
			PlaybackContext: &current.PlaybackContext.URI,
			PlaybackOffset: &spotify.PlaybackOffset{
				Position: currentTrackIndex + amt,
			},
		})
		return nil
	default:
		for i := 0; i < amt; i++ {
			client.Next(ctx)
		}
	}
	return nil
}

func Previous(ctx *gctx.Context, client *spotify.Client) error {
	err := client.Previous(ctx)
	if err != nil {
		return err
	}
	return nil
}

func Status(ctx *gctx.Context, client *spotify.Client) error {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return err
	}
	return PrintState(state)
}

func Link(ctx *gctx.Context, client *spotify.Client) (string, error) {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return "", err
	}
	return state.Item.ExternalURLs["spotify"], nil
}

func LinkContext(ctx *gctx.Context, client *spotify.Client) (string, error) {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return "", err
	}
	return string(state.PlaybackContext.ExternalURLs["spotify"]), nil
}

func NowPlaying(ctx *gctx.Context, client *spotify.Client) error {
	current, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return err
	}
	return PrintPlaying(current)
}

func Shuffle(ctx *gctx.Context, client *spotify.Client) error {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get current playstate")
	}
	err = client.Shuffle(ctx, !state.ShuffleState)
	if err != nil {
		return err
	}
	ctx.Println("Shuffle set to", !state.ShuffleState)
	return nil
}

func Repeat(ctx *gctx.Context, client *spotify.Client) error {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get current playstate")
	}
	newState := "off"
	if state.RepeatState == "off" {
		newState = "context"
	}
	// spotifyd only supports binary value for repeat, context or off, change when/if spotifyd is better
	err = client.Repeat(ctx, newState)
	if err != nil {
		return err
	}
	ctx.Println("Repeat set to", newState)
	return nil
}

func TrackList(ctx *gctx.Context, client *spotify.Client, page int) (*spotify.SavedTrackPage, error) {
	return client.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset((page-1)*50))
}

func GetQueue(ctx *gctx.Context, client *spotify.Client) (*spotify.Queue, error) {
	return client.GetQueue(ctx)
}

func Playlists(ctx *gctx.Context, client *spotify.Client, page int) (*spotify.SimplePlaylistPage, error) {
	return client.CurrentUsersPlaylists(ctx, spotify.Limit(50), spotify.Offset((page-1)*50))
}

func PlaylistTracks(ctx *gctx.Context, client *spotify.Client, playlist spotify.ID, page int) (*spotify.PlaylistTrackPage, error) {
	return client.GetPlaylistTracks(ctx, playlist, spotify.Limit(50), spotify.Offset((page-1)*50))
}

func PrintState(state *spotify.PlayerState) error {
	state.Item.AvailableMarkets = []string{}
	state.Item.Album.AvailableMarkets = []string{}
	out, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func PrintPlaying(current *spotify.CurrentlyPlaying) error {
	icon := "▶"
	if !current.Playing {
		icon = "⏸"
	}
	fmt.Println(fmt.Sprintf("%s %s - %s", icon, current.Item.Name, current.Item.Artists[0].Name))
	return nil
}

func PrintDevices(devices []spotify.PlayerDevice) error {
	out, err := json.MarshalIndent(devices, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func SetDevice(ctx *gctx.Context, client *spotify.Client, device spotify.PlayerDevice) error {
	out, err := json.MarshalIndent(device, "", " ")
	if err != nil {
		return err
	}
	configDir, _ := os.UserConfigDir()
	err = os.WriteFile(filepath.Join(configDir, "gospt/device.json"), out, 0o644)
	if err != nil {
		return err
	}
	_, err = activateDevice(ctx, client)
	if err != nil {
		return err
	}
	return nil
}

func isNoActiveError(err error) bool {
	return strings.Contains(err.Error(), "No active device found")
}

func RadioFromPlaylist(ctx *gctx.Context, client *spotify.Client, playlist spotify.SimplePlaylist) error {
	rand.Seed(time.Now().Unix())
	total := playlist.Tracks.Total
	if total == 0 {
		return fmt.Errorf("This playlist is empty")
	}
	pages := int(math.Ceil(float64(total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(int(pages-1)) + 1
	}
	playlistPage, err := client.GetPlaylistItems(ctx, playlist.ID, spotify.Limit(50), spotify.Offset((randomPage-1)*50))
	if err != nil {
		return err
	}
	pageSongs := playlistPage.Items
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 5
	if len(pageSongs) < seedCount {
		seedCount = len(pageSongs)
	}
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.Track.Track.ID)
	}
	return RadioGivenList(ctx, client, seedIds[:seedCount], playlist.Name)
}

func RadioFromAlbum(ctx *gctx.Context, client *spotify.Client, album spotify.SimpleAlbum) error {
	rand.Seed(time.Now().Unix())
	tracks, err := AlbumTracks(ctx, client, album.ID, 1)
	if err != nil {
		return err
	}
	total := tracks.Total
	if total == 0 {
		return fmt.Errorf("This playlist is empty")
	}
	pages := int(math.Ceil(float64(total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(int(pages-1)) + 1
	}
	albumTrackPage, err := AlbumTracks(ctx, client, album.ID, randomPage)
	if err != nil {
		return err
	}
	pageSongs := albumTrackPage.Tracks
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 5
	if len(pageSongs) < seedCount {
		seedCount = len(pageSongs)
	}
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.ID)
	}
	return RadioGivenList(ctx, client, seedIds[:seedCount], album.Name)
}

func RadioFromSavedTracks(ctx *gctx.Context, client *spotify.Client) error {
	rand.Seed(time.Now().Unix())
	savedSongs, err := client.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset(0))
	if err != nil {
		return err
	}
	if savedSongs.Total == 0 {
		return fmt.Errorf("You have no saved songs")
	}
	pages := int(math.Ceil(float64(savedSongs.Total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(int(pages-1)) + 1
	}
	trackPage, err := client.CurrentUsersTracks(ctx, spotify.Limit(50), spotify.Offset(randomPage*50))
	if err != nil {
		return err
	}
	pageSongs := trackPage.Tracks
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 4
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.ID)
	}
	seedIds = append(seedIds, savedSongs.Tracks[0].ID)
	return RadioGivenList(ctx, client, seedIds, "Saved Tracks")
}

func RadioGivenList(ctx *gctx.Context, client *spotify.Client, song_ids []spotify.ID, name string) error {
	seed := spotify.Seeds{
		Tracks: song_ids,
	}
	recomendations, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(99))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	err = ClearRadio(ctx, client)
	if err != nil {
		return err
	}
	radioPlaylist, err := GetRadioPlaylist(ctx, client, name)
	if err != nil {
		return err
	}
	queue := []spotify.ID{song_ids[0]}
	all_recs := map[spotify.ID]bool{}
	all_recs[song_ids[0]] = true
	for _, rec := range recomendationIds {
		all_recs[rec] = true
		queue = append(queue, rec)
	}
	_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	err = client.PlayOpt(ctx, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
		PlaybackOffset: &spotify.PlaybackOffset{
			Position: 0,
		},
	})
	if err != nil {
		if isNoActiveError(err) {
			deviceId, err := activateDevice(ctx, client)
			if err != nil {
				return err
			}
			err = client.PlayOpt(ctx, &spotify.PlayOptions{
				PlaybackContext: &radioPlaylist.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: 0,
				},
				DeviceID: &deviceId,
			})
			if err != nil {
				return err
			}
		}
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additional_recs, err := client.GetRecommendations(ctx, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additional_recs.Tracks {
			if _, ok := all_recs[song.ID]; !ok {
				all_recs[song.ID] = true
				additionalRecsIds = append(additionalRecsIds, song.ID)
				queue = append(queue, song.ID)
			}
		}
		_, err = client.AddTracksToPlaylist(ctx, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}

func activateDevice(ctx *gctx.Context, client *spotify.Client) (spotify.ID, error) {
	var device *spotify.PlayerDevice
	configDir, _ := os.UserConfigDir()
	if _, err := os.Stat(filepath.Join(configDir, "gospt/device.json")); err == nil {
		deviceFile, err := os.Open(filepath.Join(configDir, "gospt/device.json"))
		if err != nil {
			return "", err
		}
		defer deviceFile.Close()
		deviceValue, err := io.ReadAll(deviceFile)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(deviceValue, &device)
		if err != nil {
			return "", err
		}
		err = client.TransferPlayback(ctx, device.ID, true)
		if err != nil {
			return "", err
		}
	} else {
		fmt.Println("YOU MUST RUN gospt setdevice FIRST")
	}
	return device.ID, nil
}

func getDefaultDevice(ctx *gctx.Context, client *spotify.Client) (spotify.ID, error) {
	configDir, _ := os.UserConfigDir()
	if _, err := os.Stat(filepath.Join(configDir, "gospt/device.json")); err == nil {
		deviceFile, err := os.Open(filepath.Join(configDir, "gospt/device.json"))
		if err != nil {
			return "", err
		}
		defer deviceFile.Close()
		deviceValue, err := io.ReadAll(deviceFile)
		if err != nil {
			return "", err
		}
		var device *spotify.PlayerDevice
		err = json.Unmarshal(deviceValue, &device)
		if err != nil {
			return "", err
		}
		return device.ID, nil
	} else {
		return "", err
	}
}

func GetRadioPlaylist(ctx *gctx.Context, client *spotify.Client, name string) (*spotify.FullPlaylist, error) {
	configDir, _ := os.UserConfigDir()
	playlistFile, err := os.ReadFile(filepath.Join(configDir, "gospt/radio.json"))
	if errors.Is(err, os.ErrNotExist) {
		return CreateRadioPlaylist(ctx, client, name)
	}
	if err != nil {
		return nil, err
	}
	var playlist *spotify.FullPlaylist
	err = json.Unmarshal(playlistFile, &playlist)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

func CreateRadioPlaylist(ctx *gctx.Context, client *spotify.Client, name string) (*spotify.FullPlaylist, error) {
	// private flag doesnt work
	configDir, _ := os.UserConfigDir()
	playlist, err := client.CreatePlaylistForUser(ctx, ctx.UserId, name+" - autoradio", "Automanaged radio playlist", false, false)
	if err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(playlist, "", " ")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(filepath.Join(configDir, "gospt/radio.json"), raw, 0o644)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}
