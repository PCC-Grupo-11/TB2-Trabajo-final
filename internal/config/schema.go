package config

const (
	NumericFeatures = 7

	AgencyCount     = 14
	ComplaintCount  = 110
	DescriptorCount = 397
	LocationCount   = 41
	BoroughCount    = 6
	ComboCount      = 522

	AgencyOffset     = NumericFeatures
	ComplaintOffset  = AgencyOffset + AgencyCount
	DescriptorOffset = ComplaintOffset + ComplaintCount
	LocationOffset   = DescriptorOffset + DescriptorCount
	BoroughOffset    = LocationOffset + LocationCount
	ComboOffset      = BoroughOffset + BoroughCount
	TotalFeatures    = ComboOffset + ComboCount

	NumClasses = 8
)
