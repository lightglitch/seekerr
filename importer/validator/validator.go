/*
 * Copyright © 2020 Mário Franco
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package validator

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/lightglitch/seekerr/provider"
	"github.com/rs/zerolog"
	"time"
)

type RuleEnv struct {
	provider.ListItem
	Now func() time.Time
}

func NewRuleValidatior(logger *zerolog.Logger) *RuleValidatior {
	return &RuleValidatior{
		logger: logger.With().Str("Component", "Rule Validator").Logger(),
		rules:  nil,
	}
}

type RuleValidatior struct {
	logger        zerolog.Logger
	rules         []*vm.Program
	revisionRules []*vm.Program
}

func (v *RuleValidatior) InitRules(config provider.ListConfig) error {
	v.rules = []*vm.Program{}
	v.revisionRules = []*vm.Program{}
	env := &RuleEnv{}

	v.logger.Debug().Interface("rules", config.Filter.Exclude).Msg("Prepare list rules")
	for _, rule := range config.Filter.Exclude {
		compiledRule, err := expr.Compile(rule, expr.Env(env), expr.AsBool())
		if err != nil {
			v.logger.Error().Err(err).Msgf("Invalid exclude rule: %q", rule)
			return err
		}

		v.rules = append(v.rules, compiledRule)
	}

	v.logger.Debug().Int("rules", len(v.rules)).Msg("Initialized list rules")

	v.logger.Debug().Interface("revision rules", config.Filter.Revision).Msg("Prepare list rules")
	for _, rule := range config.Filter.Revision {
		compiledRule, err := expr.Compile(rule, expr.Env(env), expr.AsBool())
		if err != nil {
			v.logger.Error().Err(err).Msgf("Invalid revision rule: %q", rule)
			return err
		}

		v.revisionRules = append(v.revisionRules, compiledRule)
	}

	v.logger.Debug().Int("revision rules", len(v.revisionRules)).Msg("Initialized list rules")

	return nil
}

func (v *RuleValidatior) IsItemForRevision(item *provider.ListItem) bool {

	env := RuleEnv{
		ListItem: *item,
		Now:      func() time.Time { return time.Now().UTC() },
	}

	v.logger.Debug().Int("rules", len(v.revisionRules)).Msg("Validating item rules")

	for _, rule := range v.revisionRules {
		result, err := expr.Run(rule, env)
		if err != nil {
			v.logger.Error().Err(err).Interface("item", item).Msg("Failed validation rule for item")
			return false
		}

		expResult, ok := result.(bool)
		if !ok || expResult {
			return false
		}
	}

	v.logger.Debug().Msg("Item approved for revision")

	return true
}

func (v *RuleValidatior) IsItemApproved(item *provider.ListItem) bool {

	env := RuleEnv{
		ListItem: *item,
		Now:      func() time.Time { return time.Now().UTC() },
	}

	v.logger.Debug().Int("rules", len(v.rules)).Msg("Validating item rules")

	for _, rule := range v.rules {
		result, err := expr.Run(rule, env)
		if err != nil {
			v.logger.Error().Err(err).Interface("item", item).Msg("Failed validation rule for item")
			return false
		}

		expResult, ok := result.(bool)
		if !ok || expResult {
			return false
		}
	}

	v.logger.Debug().Msg("Item approved")

	return true
}
