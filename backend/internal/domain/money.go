package domain

import (
	"fmt"
	"strconv"
	"strings"
)

const maxMoneyRubles = 1_000_000_000_000

type Money struct {
	cents int64
}

func NewMoney(cents int64) Money {
	return Money{cents: cents}
}

func (m Money) GetCents() int64 {
	return m.cents
}

func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

func (m Money) Sub(other Money) Money {
	return Money{cents: m.cents - other.cents}
}

func (m Money) MulInt(mul int64) Money {
	return Money{cents: m.cents * mul}
}

func (m Money) DivInt(div int64) (Money, error) {
	if div <= 0 {
		return Money{}, NewDomainError(
			ErrorDivisionByZero,
			"divider",
			"divider must be greater than 0",
		)
	}
	return Money{cents: ((m.GetCents() + div/2) / div)}, nil
}

func (m Money) Abs() Money {
	if m.cents < 0 {
		return Money{cents: -m.cents}
	}
	return m
}

func (m Money) IsPositive() bool {
	return m.cents > 0
}

func (m Money) GreaterOrEqual(other Money) bool {
	return m.cents >= other.cents
}

func ParseMoney(value string) (Money, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, ",", ".")
	if value == "" {
		return Money{}, NewDomainError(ErrorEmptyMoney, "money", "money value is empty")
	}
	if strings.HasPrefix(value, "-") {
		return Money{}, NewDomainError(ErrorNegativeMoney, "money", "money value must not be negative")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return Money{}, NewDomainError(ErrorInvalidMoneyFormat, "money", "money must contain only one decimal separator")
	}
	rubles, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Money{}, NewDomainError(ErrorInvalidMoneyFormat, "money", "invalid rubles value")
	}
	var cents int64
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) > 2 {
			return Money{}, NewDomainError(ErrorMoneyTooManyDecimalDigits, "money", "money must have no more than 2 decimal digits")
		}
		if fraction == "" {
			fraction = "0"
		}
		if len(fraction) == 1 {
			fraction += "0"
		}
		cents, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return Money{}, NewDomainError(ErrorInvalidMoneyFormat, "money", "invalid cents value")
		}
	}
	if rubles > maxMoneyRubles {
		return Money{}, NewDomainError(ErrorInvalidMoneyFormat, "money", "money value is too large")
	}
	total := rubles*100 + cents
	return Money{cents: total}, nil
}

func (m Money) String() string {
	sign := ""
	cents := m.cents
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}
