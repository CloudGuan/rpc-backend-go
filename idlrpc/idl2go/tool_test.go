package main

import "testing"

func TestDealPbStructField(t *testing.T) {
	idlCases := []struct {
		input  string
		except string
	}{
		{"skill", "Skill"},
		{"skill_a", "SkillA"},
		{"skill_", "Skill_"},
		{"_skill", "XSkill"},
		{"skill_ab_bb", "SkillAbBb"},
		{"skill_a_1", "SkillA_1"},
	}

	for _, tCase := range idlCases {
		res := DealPbStructField(tCase.input)
		if res != tCase.except {
			t.Errorf("%s no except %s", res, tCase.except)
		}
	}
}
