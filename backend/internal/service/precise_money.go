package service

import "lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"

type preciseMoney struct {
	numeratorCents int64
	denominator    int64
}

func zeroPreciseMoney() preciseMoney {
	return preciseMoney{
		numeratorCents: 0,
		denominator:    1,
	}
}

func gcd(a int64, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func (m preciseMoney) normalize() preciseMoney {
	if m.denominator < 0 {
		m.numeratorCents = -m.numeratorCents
		m.denominator = -m.denominator
	}
	if m.numeratorCents == 0 {
		return zeroPreciseMoney()
	}
	divisor := gcd(abs(m.numeratorCents), m.denominator)
	return preciseMoney{
		numeratorCents: m.numeratorCents / divisor,
		denominator:    m.denominator / divisor,
	}
}

func roundPositiveFraction(numerator int64, denominator int64) int64 {
	return (numerator + denominator/2) / denominator
}

func (m preciseMoney) RoundMoney() domain.Money {
	numerator := m.numeratorCents
	if numerator < 0 {
		return domain.NewMoney(-roundPositiveFraction(-numerator, m.denominator))
	}
	return domain.NewMoney(roundPositiveFraction(numerator, m.denominator))
}

func divideMoneyPrecisely(value domain.Money, divider int64) (preciseMoney, error) {
	if divider <= 0 {
		return preciseMoney{}, domain.NewDomainError(
			domain.ErrorDivisionByZero,
			"divider",
			"divider must be greater than 0",
		)
	}
	return preciseMoney{
		numeratorCents: value.GetCents(),
		denominator:    divider,
	}.normalize(), nil
}

func (m preciseMoney) DivInt(divider int64) (preciseMoney, error) {
	if divider <= 0 {
		return preciseMoney{}, domain.NewDomainError(
			domain.ErrorDivisionByZero,
			"divider",
			"divider must be greater than 0",
		)
	}
	return preciseMoney{
		numeratorCents: m.numeratorCents,
		denominator:    m.denominator * divider,
	}.normalize(), nil
}

func (m preciseMoney) Add(other preciseMoney) preciseMoney {
	return preciseMoney{
		numeratorCents: m.numeratorCents*other.denominator + other.numeratorCents*m.denominator,
		denominator:    m.denominator * other.denominator,
	}.normalize()
}
