package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"polling-service/internal"
)

type Vote struct {
	Option string `json:"option"`
}

type VoteHandler struct {
	voteCaster internal.VoteCaster
}

func NewVoteHandler(voteCaster internal.VoteCaster) *VoteHandler {
	return &VoteHandler{
		voteCaster: voteCaster,
	}
}

func (h *VoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var vote Vote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if vote.Option == "" {
		http.Error(w, "Vote option cannot be empty", http.StatusBadRequest)
		return
	}

	if err := h.voteCaster.CastVote(vote.Option); err != nil {
		log.Printf("Failed to cast vote: %s", err)
		http.Error(w, "Failed to process vote", http.StatusInternalServerError)
		return
	}

	log.Printf("Published vote for option: %s", vote.Option)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Vote cast successfully!\n"))
}
