package tc

import (
	"fmt"
	"math"
	"time"
)

/*
 * Rate algorithm interface
 * - base on `github.com/alextanhongpin/go-rate`
 */

 //rate info
 type Rate struct {
 }

//construct
func NewRate() *Rate {
	this := &Rate{}
	return this
}

 //testing
func (r *Rate) Testing() {
	var (
		votes1 int64
		votes2 int64
	)
	now := time.Now()
	dd := time.Unix(now.Unix(), 0)
	fmt.Println("now:", now)
	fmt.Println("dd:", dd)

	votes1 = 1
	votes2 = 1

	//used for reply score?
	w1 := r.Wilson(int(votes1), 0)
	w2 := r.Wilson(int(votes2), 0)
	diff_w := w1 - w2
	fmt.Println("w1:", w1, ", w3:", w2, ", diff_w:", diff_w, "\n")

	//
	//hh1 := r.Hacker(int(votes1), time.Unix(now.Unix(), 0))
	//hh2 := r.Hacker(int(votes2), time.Unix(now.Unix() + 1000, 0))
	//diff_hh := hh1 - hh2
	//fmt.Println("hh1:", hh1, ", hh2:", hh2, ", diff_hh:", diff_hh, "\n")
	//
	////used for praise score?
	//v1 := r.Votes(votes1, 0)
	//v2 := r.Votes(votes2, 0)
	//diff_v := v1 - v2
	//fmt.Println("v1:", v1, ", v2:", v2, ", diff_v:", diff_v, "\n")

	//used for praise + time score
	//vote value use the praise value
	//unix time use the post time value
	h1 := r.Hot(votes1, 0, time.Unix(now.Unix(), 0))
	h2 := r.Hot(votes2, 0, time.Unix(now.Unix(), 0))
	diff_h := h1 - h2
	fmt.Println("h1:", h1, ", h2:", h2, ", diff_h:", diff_h, "\n")
}

 // wilson score interval sort
 // http://www.evanmiller.org/how-not-to-sort-by-average-rating.html
 func (r *Rate) Wilson(ups, downs int) float64 {
	n := ups + downs
	if n == 0 {
		return 0
	}

	n1 := float64(n)
	// z represents the statistical confidence
	// z = 1.0 => ~69%, 1.96 => ~95% (default)
	z := 1.96
	p := float64(ups / n)
	zzfn := z * z / (4 * n1)

	return (p + 2.0*zzfn - z*math.Sqrt((zzfn/n1+p*(1.0-p))/n1)) / (1 + 4*zzfn)
 }

 // hackernews' hot sort
 // http://amix.dk/blog/post/19574
 func (r *Rate) Hacker(votes int, date time.Time) float64 {
	gravity := 1.8
	hoursAge := float64(date.Unix() * 3600)
	return float64(votes-1) / math.Pow(hoursAge+2, gravity)
 }

 // reddit's hot sort
 // http://amix.dk/blog/post/19588
 func (r *Rate) Reddit(ups int, downs int, date time.Time) float64 {
	decay := int64(45000)
	s := float64(ups - downs)
	order := math.Log(math.Max(math.Abs(s), 1)) / math.Ln10
	return order - float64(date.Unix()/decay)
 }

 // Votes returns the score from sorting (which includes negative scores too)
 // http://nbviewer.jupyter.org/github/CamDavidsonPilon/Probabilistic-Programming-and-Bayesian-Methods-for-Hackers/blob/master/Chapter4_TheGreatestTheoremNeverTold/Ch4_LawOfLargeNumbers_PyMC3.ipynb
 func (r *Rate) Votes(upVotes, downVotes int64) float64 {
	a := float64(1 + upVotes)
	b := float64(1 + downVotes)
	mu := a / (a + b)
	stdErr := 1.65 * math.Sqrt((a*b)/(math.Pow(a+b, 2)*(a+b+1)))
	return mu - stdErr
 }

 // Stars return the scores for star ratings
 func (r *Rate) Stars(n, s int64) float64 {
	// s is sum of all the ratings
	// n is the number of users who rated
	a := float64(1 + s)
	b := float64(1 + n - s)
	mu := a / (a + b)
	stdErr := 1.65 * math.Sqrt((a*b)/(math.Pow(a+b, 2)*(a+b+1)))
	return mu - stdErr
 }

 // Hot will return ranking based on the time it is created
 func (r *Rate) Hot(upVotes, downVotes int64, date time.Time) float64 {
	s := float64(upVotes - downVotes)
	order := math.Log10(math.Max(math.Abs(s), 1))
	var sign float64
	if s > 0 {
		sign = 1.0
	} else if s < 0 {
		sign = -1.0
	} else {
		sign = 0.0
	}
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6
	// epoch_seconds := time.Date(1970,1,14, 3, 0, 28, 3e6, time.UTC).UnixNano() / 1e6
	seconds := (date.UnixNano()/1e6-epoch)/1e3 - 1134028003
	return round(sign*order+float64(seconds)/45000.0, 0.5, 7)
 }

 //inter round func
 func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
 }