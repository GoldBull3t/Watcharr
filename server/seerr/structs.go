package seerr

type StatusResponse struct {
	Version   string `json:"version"`
	CommitTag string `json:"commitTag"`
}

type MediaStatus int

const (
	MediaStatusUnknown            MediaStatus = 1
	MediaStatusPendingApproval    MediaStatus = 2
	MediaStatusProcessing         MediaStatus = 3
	MediaStatusPartiallyAvailable MediaStatus = 4
	MediaStatusAvailable          MediaStatus = 5
)

func (s MediaStatus) String() string {
	switch s {
	case MediaStatusPendingApproval:
		return "PENDING"
	case MediaStatusProcessing:
		return "PROCESSING"
	case MediaStatusPartiallyAvailable:
		return "PARTIALLY_AVAILABLE"
	case MediaStatusAvailable:
		return "AVAILABLE"
	default:
		return "UNKNOWN"
	}
}

type MediaInfo struct {
	ID             int         `json:"id,omitempty"`
	TmdbID         int         `json:"tmdbId,omitempty"`
	TvdbID         int         `json:"tvdbId,omitempty"`
	Status         MediaStatus `json:"status"`
	CreatedAt      string      `json:"createdAt,omitempty"`
	UpdatedAt      string      `json:"updatedAt,omitempty"`
	DownloadStatus []any       `json:"downloadStatus,omitempty"`
	Requests       []any       `json:"requests,omitempty"`
}

type MediaResponse struct {
	ID        int        `json:"id,omitempty"`
	MediaInfo *MediaInfo `json:"mediaInfo,omitempty"`
}

type RequestPayload struct {
	MediaType string `json:"mediaType"` // "movie" or "tv"
	MediaID   int    `json:"mediaId"`   // TMDB ID
	Seasons   any    `json:"seasons,omitempty"` // "all" or []int
}

type RequestResponse struct {
	ID     int        `json:"id"`
	Status int        `json:"status"`
	Media  *MediaInfo `json:"media,omitempty"`
}

type TestRequest struct {
	Host string `json:"host"`
	Key  string `json:"key"`
}

type ContentStatusResponse struct {
	Configured    bool        `json:"configured"`
	Status        MediaStatus `json:"status"`
	StatusName    string      `json:"statusName"`
	IsAvailable   bool        `json:"isAvailable"`
	IsProcessing  bool        `json:"isProcessing"`
	IsPending     bool        `json:"isPending"`
	MediaInfo     *MediaInfo  `json:"mediaInfo,omitempty"`
}
