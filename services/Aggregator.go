package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/seelentov/aggregator/http/req"
	"github.com/seelentov/aggregator/http/res"
	"github.com/seelentov/aggregator/models"
)

type Aggregator struct {
	url             string
	password        string
	user            string
	updateAuthMilli int64

	authToken string
	lastAuth  int64
}

func NewAggregator(url string, password string, user string, updateAuthMilli int64) (a *Aggregator, err error) {

	a = &Aggregator{url: url, password: password, user: user, updateAuthMilli: updateAuthMilli}

	err = a.auth()

	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Aggregator) auth() (err error) {
	now := time.Now().UTC().UnixMilli()

	if now-a.lastAuth >= a.updateAuthMilli || a.authToken == "" {
		var path string
		var method string

		if a.authToken == "" {
			path = "/auth"
			method = "POST"
		} else {
			path = "/refresh"
			method = "GET"
		}

		err = a.performAuthRequest(path, method)
		if err != nil {
			return err
		}
		a.lastAuth = now
	}

	return
}

func (a *Aggregator) performAuthRequest(path string, method string) error {
	resp, err := a.request(path, method, &req.AuthReq{Username: a.user, Password: a.password}, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		b := &res.AuthRes{}
		err = json.Unmarshal(bodyBytes, b)
		if err != nil {
			return err
		}
		a.authToken = b.Token
		return nil
	} else {
		return fmt.Errorf("can`t authorize to %s by %s: %v\n%s", a.url, a.user, resp.StatusCode, string(bodyBytes))
	}
}

func (a *Aggregator) request(url string, method string, body interface{}, checkAuth bool) (resp *http.Response, err error) {

	if checkAuth {
		err = a.auth()
		if err != nil {
			return nil, err
		}
	}

	var data []byte
	if body != nil {
		if reflect.TypeOf(body).String() == "string" {
			data = []byte(body.(string))
		} else {
			data, err = json.Marshal(body)
			if err != nil {
				return nil, err
			}
		}
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s/rest%s", a.url, url), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if a.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.authToken))
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	client := &http.Client{}
	resp, err = client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *Aggregator) processResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		if target != nil {
			err = json.Unmarshal(bodyBytes, target)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("request failed with status %v: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (a *Aggregator) Evaluate(req *req.EvaluateReq) (result interface{}, err error) {

	resp, err := a.request("/v1/evaluate", "POST", req, true)
	if err != nil {
		return nil, err
	}

	var res res.ResultRes[interface{}]

	err = a.processResponse(resp, &res)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (a *Aggregator) GetVariables(context string, includeFormat bool) (result []models.Variable, err error) {
	resp, err := a.request(fmt.Sprintf("/v1/contexts/%s/variables?includeFormat=%t", context, includeFormat), "GET", nil, true)
	if err != nil {
		return nil, err
	}

	err = a.processResponse(resp, &result)

	if err != nil {
		return nil, fmt.Errorf("can`t get variables %s from %s: %w", context, a.url, err)
	}
	return result, nil
}

func (a *Aggregator) GetFunctions(context string, includeFormat bool) (result []models.Function, err error) {
	resp, err := a.request(fmt.Sprintf("/v1/contexts/%s/functions?includeFormat=%t", context, includeFormat), "GET", nil, true)
	if err != nil {
		return nil, err
	}

	err = a.processResponse(resp, &result)

	if err != nil {
		return nil, fmt.Errorf("can`t get functions %s from %s: %w", context, a.url, err)
	}
	return result, nil
}

func (a *Aggregator) GetEvents(context string, includeFormat bool) (result []models.Event, err error) {
	resp, err := a.request(fmt.Sprintf("/v1/contexts/%s/events?includeFormat=%t", context, includeFormat), "GET", nil, true)
	if err != nil {
		return nil, err
	}

	err = a.processResponse(resp, &result)

	if err != nil {
		return nil, fmt.Errorf("can`t get events %s from %s: %w", context, a.url, err)
	}
	return result, nil
}

func (a *Aggregator) GetVariable(context string, variable string, limit int, offset int, target interface{}) (err error) {
	resp, err := a.request(fmt.Sprintf("/v1/contexts/%s/variables/%s?limit=%v&offset=%v", context, variable, limit, offset), "GET", nil, true)
	if err != nil {
		return err
	}

	err = a.processResponse(resp, &target)

	if err != nil {
		return fmt.Errorf("can`t get variable %s:%s from %s: %w", context, variable, a.url, err)
	}
	return nil
}

func (a *Aggregator) UpdateVariable(context string, variable string, value interface{}, method string) (err error) {

	resp, err := a.request(fmt.Sprintf("/v1/contexts/%s/variables/%s", context, variable), method, value, true)
	if err != nil {
		return err
	}

	err = a.processResponse(resp, nil)

	if err != nil {
		return fmt.Errorf("can`t update variable %s:%s from %s: %w", context, variable, a.url, err)
	}

	return
}

func (a *Aggregator) DoFunction(context string, function string, defaultTable interface{}, target interface{}) (err error) {
	if defaultTable == nil {
		defaultTable = "[]"
	}

	resp, err := a.request(fmt.Sprintf("/v1/contexts/%s/functions/%s", context, function), "POST", defaultTable, true)
	if err != nil {
		return err
	}

	err = a.processResponse(resp, &target)

	if err != nil {
		return fmt.Errorf("can`t do function %s:%s() from %s: %w", context, function, a.url, err)
	}
	return nil
}
