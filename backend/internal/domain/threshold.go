package domain

type ThresholdSource string

const (
	ThresholdSourceNone   ThresholdSource = "NONE"
	ThresholdSourceRF     ThresholdSource = "RF"
	ThresholdSourceRegion ThresholdSource = "REGION"
)

type Threshold struct {
	ValueRF     *Money
	ValueRegion *Money
	RegionCode  *string
}

type ResolvedThreshold struct {
	Value  *Money
	Source ThresholdSource
}

func (t Threshold) Resolve() ResolvedThreshold {
	rfOK := t.ValueRF != nil && t.ValueRF.IsPositive()
	regionOK := t.ValueRegion != nil && t.ValueRegion.IsPositive()

	switch {
	case !rfOK && !regionOK:
		return ResolvedThreshold{Value: nil, Source: ThresholdSourceNone}
	case rfOK && !regionOK:
		return ResolvedThreshold{Value: t.ValueRF, Source: ThresholdSourceRF}
	case !rfOK && regionOK || t.ValueRegion.GetCents() < t.ValueRF.GetCents():
		return ResolvedThreshold{Value: t.ValueRegion, Source: ThresholdSourceRegion}
	default:
		return ResolvedThreshold{Value: t.ValueRF, Source: ThresholdSourceRF}
	}
}
