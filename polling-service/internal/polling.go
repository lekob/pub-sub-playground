package internal

// VoteCaster defines the interface for casting a vote.
type VoteCaster interface {
	CastVote(option string) error
}
