package input

type AdModerationDecision string

const (
	AdModerationApprove    AdModerationDecision = "approve"
	AdModerationDisapprove AdModerationDecision = "disapprove"
)

type UpdateAdStatus struct {
	AdID     int
	Decision AdModerationDecision
}
