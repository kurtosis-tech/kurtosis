package name_generator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var singletonRandomNameGenerator *randomNameGenerator
var once sync.Once

type randomNameGenerator struct {
	random *rand.Rand
}

type generatorArgs struct {
	adjectives []string
	nouns      []string
}

// getNameGenerator we use once to seen random generator once
// https://stackoverflow.com/questions/12321133/how-to-properly-seed-random-number-generator
// has interesting read about maybe using crypto/rand in future if need be
func getNameGenerator() *randomNameGenerator {
	once.Do(func() {
		seed := time.Now().UTC().UnixNano()
		r := rand.New(rand.New(rand.NewSource(seed)))
		singletonRandomNameGenerator = &randomNameGenerator{random: r}
	})
	return singletonRandomNameGenerator
}

func (rng *randomNameGenerator) generateName(args generatorArgs) string {
	randomNoun := args.nouns[rng.random.Intn(len(args.nouns))]
	randomAdjective := args.adjectives[rng.random.Intn(len(args.adjectives))]

	randomName := fmt.Sprintf("%v-%v", randomAdjective, randomNoun)
	return randomName
}
