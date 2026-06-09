package schema

const (
	NumericFeatures = 7

	AgencyOffset = NumericFeatures
	ComplaintOffset = AgencyOffset + AgencyCount
	DescriptorOffset = ComplaintOffset + ComplaintCount
	LocationOffset = DescriptorOffset + DescriptorCount
	BoroughOffset = LocationOffset + LocationCount
	ComboOffset = BoroughOffset + BoroughCount
	TotalFeatures = ComboOffset + ComboCount
)
