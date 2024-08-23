package tela

import "fmt"

// Common detail tags
const (
	detail_Nothing           = "Nothing"
	detail_Needs_Review      = "Needs review"
	detail_Needs_Improvement = "Needs improvement"
	detail_Bugs              = "Bugs"
	detail_Errors            = "Errors"
)

// A TELA SC rating
type Rating struct {
	Address string `json:"address"` // Address of the rater
	Rating  uint64 `json:"rating"`  // The 0-99 rating number
	Height  uint64 `json:"height"`  // The block height this rating occurred
}

// TELA SC rating structure
type Rating_Result struct {
	Ratings  []Rating `json:"ratings,omitempty"` // All ratings for a SC
	Likes    uint64   `json:"likes"`             // Likes for a SC
	Dislikes uint64   `json:"dislikes"`          // Dislikes for a SC
	Average  float64  `json:"average"`           // Average category value of all ratings, will be 0-10
}

// TELA ratings variable structure
type ratings struct {
	categories      map[uint64]string
	negativeDetails map[uint64]string
	positiveDetails map[uint64]string
}

// Access TELA ratings variables and functions
var Ratings ratings

// Initialize the rating values
func initRatings() {
	Ratings.categories = map[uint64]string{
		0: "Do not use",
		1: "Broken",
		2: "Major issues",
		3: "Minor issues",
		4: "Should be improved",
		5: "Could be improved",
		6: "Average",
		7: "Good",
		8: "Very good",
		9: "Exceptional",
	}

	Ratings.negativeDetails = map[uint64]string{
		0: detail_Nothing,
		1: detail_Needs_Review,
		2: detail_Needs_Improvement,
		3: detail_Bugs,
		4: detail_Errors,
		5: "Inappropriate",
		6: "Incomplete",
		7: "Corrupted",
		8: "Plagiarized",
		9: "Malicious",
	}

	Ratings.positiveDetails = map[uint64]string{
		0: detail_Nothing,
		1: detail_Needs_Review,
		2: detail_Needs_Improvement,
		3: detail_Bugs,
		4: detail_Errors,
		5: "Visually appealing",
		6: "In depth",
		7: "Works well",
		8: "Unique",
		9: "Benevolent",
	}
}

// Returns all TELA rating categories
func (rate *ratings) Categories() (categories map[uint64]string) {
	categories = map[uint64]string{}
	for u, c := range rate.categories {
		categories[u] = c
	}

	return
}

// Returns a TELA rating category if exists, otherwise empty string
func (rate *ratings) Category(r uint64) (category string) {
	category = rate.categories[r]

	return
}

// Returns all negative TELA rating detail tags
func (rate *ratings) NegativeDetails() (details map[uint64]string) {
	details = map[uint64]string{}
	for u, c := range rate.negativeDetails {
		details[u] = c
	}

	return
}

// Returns all positive TELA rating detail tags
func (rate *ratings) PositiveDetails() (details map[uint64]string) {
	details = map[uint64]string{}
	for u, c := range rate.positiveDetails {
		details[u] = c
	}

	return
}

// Returns a TELA detail tag if exists, otherwise empty string
func (rate *ratings) Detail(r uint64, isPositive bool) (detail string) {
	if isPositive {
		detail = rate.positiveDetails[r]
	} else {
		detail = rate.negativeDetails[r]
	}

	return
}

// Parse r for its corresponding TELA rating category and detail
func (rate *ratings) Parse(r uint64) (category string, detail string, err error) {
	fP := r / 10
	sP := r % 10

	var ok bool
	if category, ok = rate.categories[fP]; !ok {
		err = fmt.Errorf("unknown category")
		return
	}

	isPositive := fP >= 5
	detail = rate.Detail(sP, isPositive)
	if detail == "" {
		err = fmt.Errorf("unknown detail")
	}

	return
}

// Parse r for its corresponding TELA rating string
func (rate *ratings) ParseString(r uint64) (rating string, err error) {
	category, detail, err := rate.Parse(r)
	if err != nil {
		return
	}

	if detail == detail_Nothing {
		rating = category
	} else {
		rating = fmt.Sprintf("%s (%s)", category, detail)
	}

	return
}

// Parse the average rating result for its category string
func (res *Rating_Result) ParseAverage() (category string) {
	category, _, _ = Ratings.Parse(uint64(res.Average))
	return
}
