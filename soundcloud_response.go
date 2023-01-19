package main

import "time"

type SoundcloudTrackInfo struct {
	ArtworkURL        string    `json:"artwork_url"`
	Caption           string    `json:"caption"`
	Commentable       bool      `json:"commentable"`
	CommentCount      int       `json:"comment_count"`
	CreatedAt         time.Time `json:"created_at"`
	Description       string    `json:"description"`
	Downloadable      bool      `json:"downloadable"`
	DownloadCount     int       `json:"download_count"`
	Duration          int       `json:"duration"`
	FullDuration      int       `json:"full_duration"`
	EmbeddableBy      string    `json:"embeddable_by"`
	Genre             string    `json:"genre"`
	HasDownloadsLeft  bool      `json:"has_downloads_left"`
	ID                int       `json:"id"`
	Kind              string    `json:"kind"`
	LabelName         string    `json:"label_name"`
	LastModified      time.Time `json:"last_modified"`
	License           string    `json:"license"`
	LikesCount        int       `json:"likes_count"`
	Permalink         string    `json:"permalink"`
	PermalinkURL      string    `json:"permalink_url"`
	PlaybackCount     int       `json:"playback_count"`
	Public            bool      `json:"public"`
	PublisherMetadata struct {
		ID            int    `json:"id"`
		Urn           string `json:"urn"`
		ContainsMusic bool   `json:"contains_music"`
	} `json:"publisher_metadata"`
	PurchaseTitle string      `json:"purchase_title"`
	PurchaseURL   string      `json:"purchase_url"`
	ReleaseDate   interface{} `json:"release_date"`
	RepostsCount  int         `json:"reposts_count"`
	SecretToken   string      `json:"secret_token"`
	Sharing       string      `json:"sharing"`
	State         string      `json:"state"`
	Streamable    bool        `json:"streamable"`
	TagList       string      `json:"tag_list"`
	Title         string      `json:"title"`
	TrackFormat   string      `json:"track_format"`
	URI           string      `json:"uri"`
	Urn           string      `json:"urn"`
	UserID        int         `json:"user_id"`
	Visuals       struct {
		Urn     string `json:"urn"`
		Enabled bool   `json:"enabled"`
		Visuals []struct {
			Urn       string `json:"urn"`
			EntryTime int    `json:"entry_time"`
			VisualURL string `json:"visual_url"`
		} `json:"visuals"`
		Tracking interface{} `json:"tracking"`
	} `json:"visuals"`
	WaveformURL string    `json:"waveform_url"`
	DisplayDate time.Time `json:"display_date"`
	Media       struct {
		Transcodings []struct {
			URL      string `json:"url"`
			Preset   string `json:"preset"`
			Duration int    `json:"duration"`
			Snipped  bool   `json:"snipped"`
			Format   struct {
				Protocol string `json:"protocol"`
				MimeType string `json:"mime_type"`
			} `json:"format"`
			Quality string `json:"quality"`
		} `json:"transcodings"`
	} `json:"media"`
	StationUrn         string `json:"station_urn"`
	StationPermalink   string `json:"station_permalink"`
	TrackAuthorization string `json:"track_authorization"`
	MonetizationModel  string `json:"monetization_model"`
	Policy             string `json:"policy"`
	User               struct {
		AvatarURL            string      `json:"avatar_url"`
		City                 string      `json:"city"`
		CommentsCount        int         `json:"comments_count"`
		CountryCode          interface{} `json:"country_code"`
		CreatedAt            time.Time   `json:"created_at"`
		CreatorSubscriptions []struct {
			Product struct {
				ID string `json:"id"`
			} `json:"product"`
		} `json:"creator_subscriptions"`
		CreatorSubscription struct {
			Product struct {
				ID string `json:"id"`
			} `json:"product"`
		} `json:"creator_subscription"`
		Description        string    `json:"description"`
		FollowersCount     int       `json:"followers_count"`
		FollowingsCount    int       `json:"followings_count"`
		FirstName          string    `json:"first_name"`
		FullName           string    `json:"full_name"`
		GroupsCount        int       `json:"groups_count"`
		ID                 int       `json:"id"`
		Kind               string    `json:"kind"`
		LastModified       time.Time `json:"last_modified"`
		LastName           string    `json:"last_name"`
		LikesCount         int       `json:"likes_count"`
		PlaylistLikesCount int       `json:"playlist_likes_count"`
		Permalink          string    `json:"permalink"`
		PermalinkURL       string    `json:"permalink_url"`
		PlaylistCount      int       `json:"playlist_count"`
		RepostsCount       int       `json:"reposts_count"`
		TrackCount         int       `json:"track_count"`
		URI                string    `json:"uri"`
		Urn                string    `json:"urn"`
		Username           string    `json:"username"`
		Verified           bool      `json:"verified"`
		Visuals            struct {
			Urn     string `json:"urn"`
			Enabled bool   `json:"enabled"`
			Visuals []struct {
				Urn       string `json:"urn"`
				EntryTime int    `json:"entry_time"`
				VisualURL string `json:"visual_url"`
			} `json:"visuals"`
			Tracking interface{} `json:"tracking"`
		} `json:"visuals"`
		Badges struct {
			Pro          bool `json:"pro"`
			ProUnlimited bool `json:"pro_unlimited"`
			Verified     bool `json:"verified"`
		} `json:"badges"`
		StationUrn       string `json:"station_urn"`
		StationPermalink string `json:"station_permalink"`
	} `json:"user"`
}
