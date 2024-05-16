package internal

import (
	"fmt"
	"math/rand"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type WeightedValueResolver struct {
	rand   Randomer
	hasher Hasher
}

func NewWeightedValueResolver(seed int64, hasher Hasher) *WeightedValueResolver {
	src := rand.NewSource(seed)
	r := rand.New(src)

	return &WeightedValueResolver{
		rand:   r,
		hasher: hasher,
	}
}

func (wve *WeightedValueResolver) Resolve(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextValueGetter) (*prefabProto.ConfigValue, int) {
	fractionThroughDistribution := wve.getUserFraction(weightedValues, propertyName, contextGetter)

	sum := int32(0)
	for _, weightedValue := range weightedValues.GetWeightedValues() {
		sum += weightedValue.GetWeight()
	}

	weightThreshold := fractionThroughDistribution * float64(sum)

	runningSum := int32(0)
	for index, weightedValue := range weightedValues.GetWeightedValues() {
		runningSum += weightedValue.GetWeight()
		if float64(runningSum) >= weightThreshold {
			return weightedValue.GetValue(), index
		}
	}

	return weightedValues.GetWeightedValues()[0].GetValue(), 0
}

func (wve *WeightedValueResolver) getUserFraction(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextValueGetter) float64 {
	if weightedValues.HashByPropertyName != nil {
		value, valueExists := contextGetter.GetContextValue(weightedValues.GetHashByPropertyName())
		if valueExists {
			valueToBeHashed := fmt.Sprintf("%s%v", propertyName, value)
			hashValue, _ := wve.hasher.HashZeroToOne(valueToBeHashed)

			return hashValue
		}
	}

	return wve.rand.Float64()
}
