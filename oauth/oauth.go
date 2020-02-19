package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/apiotrowski312/isOnline-utils-go/http_utils"
	"github.com/apiotrowski312/isOnline-utils-go/rest_errors"
)

const (
	headerXPublic   = "X-Public"
	headerXClientId = "X-Client-Id"
	headerXCallerId = "X-Caller-Id"

	paramAccessToken = "access_token"
)

type accessToken struct {
	Id       string `json:"id"`
	UserId   int64  `json:"user_id"`
	ClientId int64  `json:"client_id"`
}

func IsPublic(request *http.Request) bool {
	if request == nil {
		return true
	}
	return request.Header.Get(headerXPublic) == "true"
}

func GetCallerId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	callerId, err := strconv.ParseInt(request.Header.Get(headerXCallerId), 10, 64)

	if err != nil {
		return 0
	}

	return callerId
}

func GetClientId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	clientId, err := strconv.ParseInt(request.Header.Get(headerXClientId), 10, 64)

	if err != nil {
		return 0
	}

	return clientId
}

func AuthenticateRequest(request *http.Request) rest_errors.RestErr {
	if request == nil {
		return nil
	}

	cleanRequest(request)

	accessTokenId := strings.TrimSpace(request.URL.Query().Get(paramAccessToken))
	if accessTokenId == "" {
		return nil
	}

	at, err := getAccessToken(accessTokenId)
	if err != nil {
		if err.Status() == http.StatusNotFound {
			return nil
		}
		return err
	}

	request.Header.Add(headerXCallerId, fmt.Sprintf("%v", at.UserId))
	request.Header.Add(headerXClientId, fmt.Sprintf("%v", at.ClientId))

	return nil
}

func cleanRequest(request *http.Request) {
	if request == nil {
		return
	}

	request.Header.Del(headerXClientId)
	request.Header.Del(headerXCallerId)
}

func getAccessToken(accessTokenId string) (*accessToken, rest_errors.RestErr) {
	response, err := http_utils.Get(fmt.Sprintf("http://orchestration_oauth_1:8081/oauth/access_token/%s", accessTokenId), nil)

	if err != nil {
		return nil, rest_errors.NewInternalServerError("invalid restclient response when trying to get access token", errors.New("login error"))
	}

	bodyBytes, readErr := ioutil.ReadAll(response.Body)

	if readErr != nil {
		return nil, rest_errors.NewInternalServerError("invalid restclient response when trying to parse Body", errors.New("login error"))
	}

	if response.StatusCode > 299 {
		restErr, errBytes := rest_errors.NewRestErrorFromBytes(bodyBytes)
		fmt.Println("ERR", restErr)
		fmt.Println("ERR", errBytes)

		if errBytes != nil {
			return nil, rest_errors.NewInternalServerError("invalid error interface when try get access token", errors.New("login error"))
		}

		return nil, restErr
	}

	var at accessToken
	if err := json.Unmarshal(bodyBytes, &at); err != nil {
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal access token response", errors.New("login error"))
	}

	return &at, nil
}
