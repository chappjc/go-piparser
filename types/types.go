// Package types defines the data types needed to serialize and unserialize the
// the data sent or recieved.
package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

const (
	defaultVotesCommitMsg = "Flush vote journals"
)

var journalActionFormat, proposalToken string

// AltResponse is the possible alternative response returned if the default one
// wasn't successful.

type AltResponse struct {
	Message string `json:"message"`
	URL     string `json:"documentation_url"`
}

// HistorySHAs holds a slice of the commit history SHA token strings.
type HistorySHAs []commitSHA

// 	*AltResponse
// }

// commitSHA holds the specific commit unique SHA string value.
type commitSHA struct {
	SHA string `json:"sha"`
}

// History defines the commit full information about a commit
type History struct {
	SHA     string    `json:"sha"`
	Commit  Commit    `json:"commit"`
	URLPath string    `json:"url"`
	Files   []Content `json:"files"`
	*AltResponse
}

// Commit defines information about the committer and the commit message used.
type Commit struct {
	Committer CommitInfo `json:"committer"`
	Message   string     `json:"message"`
}

// CommitInfo defines information about the committer
type CommitInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// Content defines the changes made, filename, actual file data and other details about
// the about the commit content.
type Content struct {
	FileSHA    string `json:"sha"`
	FileName   string `json:"filename"`
	Status     string `json:"status"`
	Additions  int32  `json:"additions"`
	Deletions  int32  `json:"deletions"`
	Change     int32  `json:"changes"`
	BlobURL    string `json:"blob_url"`
	RawURL     string `json:"raw_url"`
	ContentURL string `json:"contents_url"`
	Data       *Votes `json:"patch"`
}

// Votes defines a slice type of all votes cast data.
type Votes []CastVoteData

// CastVoteData defines the struct of a cast vote and the reciept response.
type CastVoteData struct {
	*PiVote `json:"castvote"`
	Receipt string `json:"receipt"`
}

// PiVote defines the finer details about a vote.
type PiVote struct {
	Token     string `json:"token"`
	Ticket    string `json:"ticket"`
	VoteBit   string `json:"votebit"`
	Signature string `json:"signature"`
}

// SetProposalToken sets the current proposal token string whose data is being
// unmarshalled.
func SetProposalToken(token string) {
	proposalToken = token
}

// SetJournalActionFormat sets journal (struct with the version and the journal
// action) format to use for the regexp.
func SetJournalActionFormat(val string) {
	journalActionFormat = val
}

// UnmarshalJSON defines the default unmarshaller for Votes.
func (v *Votes) UnmarshalJSON(b []byte) error {
	str := string(b)
	isMatched, err := regexp.MatchString(proposalToken, str)
	if !isMatched || err != nil {
		return err
	}

	// Delete the special characters indicating addition and deletion metrics.
	str = replaceUnwanted(str, `(@{2}[\s\S]*@{2})`, "")

	// Drops github added special characters
	str, _ = strconv.Unquote(str)
	str = "[" + str + "]"

	// Replace '[ {"version":"\d","action":"add|del"}' with '['.
	str = replaceUnwanted(str, fmt.Sprintf(`([[][\s]*%s)`, journalActionFormat), "[")

	// Replace '} +{"version":"\d","action":"add|del"}' with '},'.
	str = replaceUnwanted(str, fmt.Sprintf(`(}[\s+]*%s)`, journalActionFormat), "},")

	type votes2 Votes
	var v2 votes2

	err = json.Unmarshal([]byte(str), &v2)
	if err != nil {
		return err
	}

	*v = Votes(v2)

	return nil
}

// UnmarshalJSON is the default unmarshaller for HistorySHA.
func (h *HistorySHAs) UnmarshalJSON(b []byte) error {
	isMatched, err := regexp.MatchString(defaultVotesCommitMsg, string(b))
	if !isMatched || err != nil {
		return err
	}

	type history HistorySHAs
	var h2 history

	err = json.Unmarshal(b, &h2)
	if err != nil {
		return err
	}

	*h = HistorySHAs(h2)
	return nil
}

// replaceUnwanted replaces 'x' regex expression matchings in string 'str' with 'with'.
func replaceUnwanted(str, x, with string) string {
	return regexp.MustCompile(x).ReplaceAllLiteralString(str, with)
}
