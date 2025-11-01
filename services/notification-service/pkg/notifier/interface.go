package notifier

type Notifier interface {
	NotifyStarInrcease(repo string, oldStars, newStars int) error
}
