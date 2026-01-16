package urlShortener

import "time"

type UrlDbModel struct {
	Id        int       `json:"id"`
	Url       string    `json:"url"`
	ShortCode string    `json:"shortCode"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int       `json:"-"`
}
