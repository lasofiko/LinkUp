package domain

type ReactionType string

const (
	ReactionLike    ReactionType = "like"
	ReactionDislike ReactionType = "dislike"
	ReactionBlock   ReactionType = "block"
)

type Reaction struct {
	FromUserID int
	ToUserID   int
	Type       ReactionType
}
