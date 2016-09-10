package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Dataman-Cloud/crane/src/utils/cranerror"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	baseUrl string
	server  *httptest.Server
)

const (
	CodeOk        = 0
	CodeUndefined = 10001
)

type ResponseBody struct {
	Code int   `json: "code"`
	Err  error `json:"err"`
}

func TestMain(m *testing.M) {
	server = startHttpServer()
	baseUrl = server.URL
	defer server.Close()
	os.Exit(m.Run())
}

func startHttpServer() *httptest.Server {
	router := gin.New()
	accountApi := AccountApi{
		Authenticator: NewMockAuthenticator(),
	}

	v1 := router.Group("/account/v1")
	{
		v1.POST("/groups/:group_id/account", accountApi.CreateAccount)
	}

	return httptest.NewServer(router)
}

func TestCreateAccount(t *testing.T) {
	var requestBody string = `{"email": "admin@dataman-inc.com", "password": "Dataman1234"}`

	// normal test case
	req, _ := http.NewRequest("POST", baseUrl+"/account/v1/groups/1/account", strings.NewReader(requestBody))
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "should be equal")

	// bad request test case
	var badRequestBody string = `{"emain" "admin@dataman-inc.com", "password": "Dataman1234"}`

	req, _ = http.NewRequest("POST", baseUrl+"/account/v1/groups/1/account", strings.NewReader(badRequestBody))
	resp, _ = http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, http.StatusBadRequest, "should be equal")

	// bad password test case
	var badPasswordBody string = `{"email": "admin@dataman-inc.com", "assword": "213"}`

	req, _ = http.NewRequest("POST", baseUrl+"/account/v1/groups/1/account", strings.NewReader(badPasswordBody))
	resp, _ = http.DefaultClient.Do(req)
	b, _ := ioutil.ReadAll(resp.Body)
	var response ResponseBody
	json.Unmarshal(b, &response)

	httpCode, errCode := parseError(CodeAccountCreateParamError)

	assert.Equal(t, response.Code, errCode, "should be equal")
	assert.Equal(t, resp.StatusCode, httpCode, "should be equal")

	// bad email test case
	badEmailBody := `{"emain": "admin@dataman-inc.com", "password": "213"}`

	req, _ = http.NewRequest("POST", baseUrl+"/account/v1/groups/1/account", strings.NewReader(badEmailBody))
	resp, _ = http.DefaultClient.Do(req)
	b, _ = ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &response)

	httpCode, errCode = parseError(CodeAccountCreateParamError)

	assert.Equal(t, response.Code, errCode, "should be equal")
	assert.Equal(t, resp.StatusCode, httpCode, "should be equal")

	// bad parameter test case
	req, _ = http.NewRequest("POST", baseUrl+"/account/v1/groups//account", strings.NewReader(requestBody))
	resp, _ = http.DefaultClient.Do(req)
	b, _ = ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &response)

	httpCode, errCode = parseError(CodeAccountCreateParamError)

	assert.Equal(t, response.Code, errCode, "should be equal")
	assert.Equal(t, resp.StatusCode, httpCode, "should be equal")

	// createAccount error test case
	CreateAccountError = cranerror.NewError(CodeAccountCreateAuthenticatorError, "create account error")
	defer func() {
		CreateAccountError = nil
	}()
	req, _ = http.NewRequest("POST", baseUrl+"/account/v1/groups/1/account", strings.NewReader(requestBody))
	resp, _ = http.DefaultClient.Do(req)
	b, _ = ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &response)

	httpCode, errCode = parseError(CodeAccountCreateAuthenticatorError)

	assert.Equal(t, response.Code, errCode, "should be equal")
	assert.Equal(t, resp.StatusCode, httpCode, "should be equal")
}

func parseError(errorCode string) (int, int) {
	httpCode := http.StatusServiceUnavailable
	errCode := CodeUndefined

	codes := strings.Split(errorCode, "-")
	if len(codes) == 2 {
		httpCode, _ = strconv.Atoi(codes[0])
		errCode, _ = strconv.Atoi(codes[1])
	}

	return httpCode, errCode
}