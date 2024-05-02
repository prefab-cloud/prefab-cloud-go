package prefab

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

func (wve *WeightedValueResolver) Resolve(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextGetter) (valueResult *prefabProto.ConfigValue, index int) {
	fractionThroughDistribution := wve.getUserFraction(weightedValues, propertyName, contextGetter)
	sum := int32(0)
	for _, weightedValue := range weightedValues.WeightedValues {
		sum += weightedValue.Weight
	}

	weightThreshold := fractionThroughDistribution * float64(sum)

	runningSum := int32(0)
	for index, weightedValue := range weightedValues.WeightedValues {
		runningSum += weightedValue.Weight
		if float64(runningSum) >= weightThreshold {
			return weightedValue.Value, index
		}
	}

	return weightedValues.WeightedValues[0].Value, 0
}

func (wve *WeightedValueResolver) getUserFraction(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextGetter) float64 {
	if weightedValues.HashByPropertyName != nil {
		value, valueExists := contextGetter.GetValue(*weightedValues.HashByPropertyName)
		if valueExists {
			valueToBeHashed := fmt.Sprintf("%s%v", propertyName, value)
			hashValue, _ := wve.hasher.HashZeroToOne(valueToBeHashed)
			return hashValue
		}
	}
	return wve.rand.Float64()
}
