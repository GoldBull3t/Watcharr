package domain

type MediaVideoType string

const (
	MediaVideoTypeTrailer         MediaVideoType = "trailer"
	MediaVideoTypeTeaser          MediaVideoType = "teaser"
	MediaVideoTypeClip            MediaVideoType = "clip"
	MediaVideoTypeFeaturette      MediaVideoType = "featurette"
	MediaVideoTypeBehindTheScenes MediaVideoType = "behind_the_scenes"
	MediaVideoTypeBloopers        MediaVideoType = "bloopers"
	MediaVideoTypeOpeningCredits  MediaVideoType = "opening_credits"
	MediaVideoTypeOther           MediaVideoType = "other"
)

// A video (trailer, etc).
// Only supports YouTube for now, but if needed we can add a `site` attribute later.
type MediaVideo struct {
	// Video ID for external platform.
	ID string `json:"id,omitempty"`
	// Video Name.
	Name string `json:"name,omitempty"`
	// Type of video.
	Type MediaVideoType `json:"type,omitempty"`
	// If this is the best video to present to the user first.
	// (eg with our View Trailer button).
	Best bool `json:"best,omitempty"`
}
