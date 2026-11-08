package domain

type Quarter int

const (
	Quarter1 Quarter = 1
	Quarter2 Quarter = 2
	Quarter3 Quarter = 3
	Quarter4 Quarter = 4

	FirstYear  int = 2021
	CountMonth int = 3
)

type Period struct {
	Year    int
	Quarter Quarter
}

type MonthMeta struct {
	Index       int
	MonthNumber int
}

func (p Period) Validate() error {
	if p.Year < FirstYear {
		return NewDomainError(ErrorInvalidYear, "period.year", "year must be greater than or equal to 2021")
	}

	if p.Quarter < Quarter1 || p.Quarter > Quarter4 {
		return NewDomainError(ErrorInvalidQuarter, "period.quarter", "quarter must be between 1 and 4")
	}
	return nil
}

func (p Period) PreviousQuarter() Period {
	if p.Quarter == Quarter1 {
		return Period{Year: p.Year - 1, Quarter: Quarter4}
	}
	return Period{Year: p.Year, Quarter: p.Quarter - 1}
}

func (p Period) Months() [CountMonth]MonthMeta {
	startMonth := (int(p.Quarter)-1)*3 + 1
	var months [CountMonth]MonthMeta
	for i := 0; i < CountMonth; i++ {
		months[i] = MonthMeta{
			Index:       i + 1,
			MonthNumber: startMonth + i,
		}
	}
	return months
}
