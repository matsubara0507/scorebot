package scorebot

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sort"
)

type Challenge struct {
	Title   string   `yaml:"title"`
	Detail  string   `yaml:"detail"`
	Point   int      `yaml:"point"`
	Choices []string `yaml:"choices"`
	Flag    string
}

type Challenges map[string]Challenge

func (challenge Challenge) SetFlag(flag string) Challenge {
	challenge.Flag = flag
	return challenge
}

func ReadChallengesYaml(challengesYaml string) (Challenges, error) {
	buf, err := ioutil.ReadFile(challengesYaml)
	if err != nil {
		return nil, err
	}

	challenges := make(map[string]Challenge)
	err = yaml.Unmarshal(buf, &challenges)
	if err != nil {
		return nil, err
	}
	return challenges, nil
}

func (challenges Challenges) Keys() []string {
	var keys []string
	for k, _ := range challenges {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (challenges Challenges) Submit(cond func(submit string, flag string) bool, flag string) (bool, string) {
	for k, v := range challenges {
		if cond(flag, v.Flag) {
			return true, k
		}
	}
	return false, ""
}

type ChallengeTable struct {
	FindAll func(challenges *Challenges) error
}
