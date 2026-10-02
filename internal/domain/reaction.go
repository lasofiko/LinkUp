package domain

type ReactionType string

const (
	ReactionLike ReactionType = "like"
)

type Reaction struct {
	FromUserID int
	ToUserID   int
	Type       ReactionType
}

func NewReaction(from, to int, reactionType ReactionType) (Reaction, error) {
	if from == to {
		return Reaction{}, ErrSelfReaction
	}

	if reactionType != ReactionLike {
		return Reaction{}, ErrInvalidReactionType
	}

	return Reaction{
		FromUserID: from,
		ToUserID:   to,
		Type:       reactionType,
	}, nil
}
