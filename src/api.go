package qj

import (
  "errors"
  "fmt"
  "strings"
  "net/http"
  "net/url"
  "io/ioutil"
  "encoding/base64"
  "encoding/json"
)

type JiraApiContext struct {
  Email   string
  ApiKey  string
  BaseUrl string
}

type JiraIssue struct {
  Id  string
  Key string
  Fields struct {
    Assignee  JiraAssignee
    Summary   string
    Labels    []string
  }
}

type JiraAssignee struct {
  EmailAddress  string
  DisplayName   string
}

type searchResults struct {
  Total   int
  Issues  []JiraIssue
}

func prepareRequest(ctx JiraApiContext, req *http.Request) {
  key := []byte(ctx.Email + ":" + ctx.ApiKey)
  encodedKey := base64.StdEncoding.EncodeToString(key)
  req.Header.Add("Authorization", "Basic " + encodedKey)
}

func apiUrl(ctx JiraApiContext, url string) string {
  if !strings.HasPrefix(url, "/") && !strings.HasSuffix(ctx.BaseUrl, "/") {
    return ctx.BaseUrl + "/" + url
  }

  return ctx.BaseUrl + url
}

func FetchIssuesJql(ctx JiraApiContext, jql string) ([]JiraIssue, error) {
  client := &http.Client{}

  params := url.Values{}
  params.Add("jql", jql)
  params.Add("maxResults", "10")

  req, err := http.NewRequest("GET", apiUrl(ctx, "rest/api/2/search") + "?" + params.Encode(), nil)
  if err != nil {
    panic(err)
  }

  prepareRequest(ctx, req)

  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }

  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(err)
  }

  if resp.StatusCode != 200 {
    return nil, errors.New(fmt.Sprintf("JIRA API call failed: %s\n", body))
  }

  results := searchResults{} 
  err = json.Unmarshal(body, &results)
  if err != nil {
    panic(err)
  }

  return results.Issues, nil
}
