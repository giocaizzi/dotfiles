package usage

import "errors"

// ParseAnthropicProfile parses GET api.anthropic.com/api/oauth/profile. It
// fails when the account uuid is missing.
func ParseAnthropicProfile(body []byte) (Profile, error) {
	_ = body
	return Profile{}, errors.New("not implemented")
}

// ParseGitHubUser parses GET api.github.com/user. It fails when the login is
// missing.
func ParseGitHubUser(body []byte) (Profile, error) {
	_ = body
	return Profile{}, errors.New("not implemented")
}
