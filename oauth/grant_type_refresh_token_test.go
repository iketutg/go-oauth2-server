package oauth

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestTestRefreshTokenGrantScopeCannotBeGreater() {
	// Insert a test refresh token
	if err := suite.db.Create(&RefreshToken{
		Token:     "test_token",
		ExpiresAt: time.Now().Add(+10 * time.Second),
		Client:    suite.client,
		User:      suite.user,
		Scope:     "foo bar",
	}).Error; err != nil {
		log.Fatal(err)
	}

	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/something", nil)
	if err != nil {
		log.Fatal(err)
	}
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"test_token"},
		"scope":         {"foo bar qux"},
	}

	w := httptest.NewRecorder()
	suite.service.refreshTokenGrant(w, r, suite.client)

	// Check the status code
	assert.Equal(suite.T(), 400, w.Code)

	// Check the response body
	assert.Equal(
		suite.T(), "{\"error\":\"Requested scope cannot be greater\"}",
		strings.TrimSpace(w.Body.String()),
	)
}

func (suite *OauthTestSuite) TestRefreshTokenGrant() {
	// Insert a test refresh token
	if err := suite.db.Create(&RefreshToken{
		Token:     "test_token",
		ExpiresAt: time.Now().Add(+10 * time.Second),
		Client:    suite.client,
		User:      suite.user,
		Scope:     "foo bar",
	}).Error; err != nil {
		log.Fatal(err)
	}

	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/something", nil)
	if err != nil {
		log.Fatal(err)
	}
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"test_token"},
		"scope":         {"foo bar"},
	}

	w := httptest.NewRecorder()
	suite.service.refreshTokenGrant(w, r, suite.client)

	// Check the status code
	assert.Equal(suite.T(), 200, w.Code)

	// Check the correct data was inserted
	accessToken := new(AccessToken)
	assert.False(suite.T(), suite.db.First(accessToken).RecordNotFound())

	// Check the response body
	expected, _ := json.Marshal(map[string]interface{}{
		"id":            accessToken.ID,
		"access_token":  accessToken.Token,
		"expires_in":    3600,
		"token_type":    "Bearer",
		"scope":         "foo bar",
		"refresh_token": "test_token",
	})
	assert.Equal(suite.T(), string(expected), strings.TrimSpace(w.Body.String()))
}