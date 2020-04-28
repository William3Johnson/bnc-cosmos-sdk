package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math"
	"math/big"
)

type Sharer struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
}

type Reward struct {
	AccAddr sdk.AccAddress
	Amount  int64
}

func allocate(sharers []Sharer, totalRewards sdk.Dec, totalShares sdk.Dec) (rewards []Reward) {
	var minToDistribute int64
	var shouldCarry []Reward
	var shouldNotCarry []Reward
	for _, sharer := range sharers {

		afterRoundDown, firstDecimalValue := mulQuoDecWithExtraDecimal(sharer.Shares, totalRewards, totalShares, 1)

		if firstDecimalValue < threshold {
			shouldNotCarry = append(shouldNotCarry, Reward{sharer.AccAddr, afterRoundDown})
		} else {
			shouldCarry = append(shouldCarry, Reward{sharer.AccAddr, afterRoundDown})
		}
		minToDistribute += afterRoundDown
	}
	remainingRewards := totalRewards.RawInt() - minToDistribute
	if remainingRewards > 0 {
		for i := range shouldCarry {
			if remainingRewards <= 0 {
				break
			} else {
				shouldCarry[i].Amount++
				remainingRewards--
			}
		}
		if remainingRewards > 0 {
			for i := range shouldNotCarry {
				if remainingRewards <= 0 {
					break
				} else {
					shouldNotCarry[i].Amount++
					remainingRewards--
				}
			}
		}
	}

	return append(shouldCarry, shouldNotCarry...)
}

// calculate a * b / c, getting the extra decimal digital as result of extraDecimalValue. For example:
// 0.00000003 * 2 / 0.00000004 = 0.000000015,
// assume that decimal place number of Dec is 8, and the extraDecimalPlace was given 1, then
// we take the 8th decimal place value '1' as afterRoundDown, and extra decimal value(9th) '5' as extraDecimalValue
func mulQuoDecWithExtraDecimal(a, b, c sdk.Dec, extraDecimalPlace int) (afterRoundDown int64, extraDecimalValue int) {
	extra := int64(math.Pow(10, float64(extraDecimalPlace)))
	product, ok := sdk.Mul64(a.RawInt(), b.RawInt())
	if !ok { // int64 exceed
		return mulQuoBigIntWithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
	} else {
		if product, ok = sdk.Mul64(product, extra); !ok {
			return mulQuoBigIntWithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
		}
		resultOfAddDecimalPlace := product / c.RawInt()
		afterRoundDown = resultOfAddDecimalPlace / extra
		extraDecimalValue = int(resultOfAddDecimalPlace % extra)
		return afterRoundDown, extraDecimalValue
	}
}

func mulQuoBigIntWithExtraDecimal(a, b, c, extra *big.Int) (afterRoundDown int64, extraDecimalValue int) {
	product := sdk.MulBigInt(sdk.MulBigInt(a, b), extra)
	result := sdk.QuoBigInt(product, c)

	expectedDecimalValueBig := &big.Int{}
	afterRoundDownBig, expectedDecimalValueBig := result.QuoRem(result, extra, expectedDecimalValueBig)
	afterRoundDown = afterRoundDownBig.Int64()
	extraDecimalValue = int(expectedDecimalValueBig.Int64())
	return afterRoundDown, extraDecimalValue
}
