package name_generator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var randomNameGenerator *RandomNameGenerator
var once sync.Once

type RandomNameGenerator struct {
	random *rand.Rand
}

type GeneratorArgs struct {
	adjectives []string
	nouns      []string
}

// GetNameGenerator we use once to seen random generator once
// https://stackoverflow.com/questions/12321133/how-to-properly-seed-random-number-generator
// has interesting read about maybe using crypto/rand in future if need be
func GetNameGenerator() *RandomNameGenerator {
	if randomNameGenerator == nil {
		once.Do(func() {
			seed := time.Now().UTC().UnixNano()
			r := rand.New(rand.New(rand.NewSource(seed)))
			randomNameGenerator = &RandomNameGenerator{random: r}
		})
	}
	return randomNameGenerator
}

func (rng *RandomNameGenerator) GenerateName(args GeneratorArgs) string {
	randomNoun := args.nouns[rng.random.Intn(len(args.nouns))]
	randomAdjective := args.adjectives[rng.random.Intn(len(args.adjectives))]

	randomName := fmt.Sprintf("%v-%v", randomAdjective, randomNoun)
	return randomName
}
