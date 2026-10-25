package domain

type (
	TagAddRequest struct {
		Name        string `json:"name" binding:"required"`
		Color       string `json:"color"`
		BgColor     string `json:"bgColor"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"isPublic,omitempty"`
	}

	TagVisibilityRequest struct {
		IsPublic bool `json:"isPublic"`
	}

	PublicTagResponse struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Color       string  `json:"color"`
		BgColor     string  `json:"bgColor"`
		Description string  `json:"description,omitempty"`
		IsPublic    bool    `json:"isPublic"`
		AuthorName  string  `json:"authorName"`
		TotalItems  int     `json:"totalItems"`
		Items       []Media `json:"items"`
	}
)
