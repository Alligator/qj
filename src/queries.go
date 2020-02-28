package qj

import (
  "os"
  "fmt"
  "strings"
  "errors"
  "io/ioutil"
  "path/filepath"
  "github.com/BurntSushi/toml"
)

type Query struct {
  Name  string
  Jql   string
}

type ConfigFile struct {
  Email   string
  ApiKey  string
  BaseUrl string
  Queries []Query
}

var exampleConfig = ConfigFile {
  Email: "",
  ApiKey: "",
  BaseUrl: "",
  Queries: []Query {
    Query {
      Name: "Assigned to me",
      Jql: "assignee = \"alligator\"",
    },
    Query {
      Name: "Recently updated",
      Jql: "order by updated",
    },
  },
}

func ConfigPath() string {
  configDir, err := os.UserConfigDir()
  if err != nil {
    panic(err)
  }

  return filepath.Join(configDir, "qj", "config.toml")
}

func ensureConfigFile() error {
  cfgPath := ConfigPath()
  dirPath := filepath.Dir(cfgPath)

  _, err := os.Stat(dirPath)
  if os.IsNotExist(err) {
    os.Mkdir(dirPath, 0666)
  } else if err != nil {
    panic(err)
  }

  file, err := os.OpenFile(cfgPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
  if err == nil {
    // file was created without error, it's new
    defer file.Close()

    err = toml.NewEncoder(file).Encode(exampleConfig)
    if err != nil {
      panic(err)
    }
    return errors.New(fmt.Sprintf("config file not found, created a default in %s\n", cfgPath))
  } else if os.IsExist(err) {
    // already exists, nothing to do
    return nil
  } else {
    panic(err)
  }

  return nil
}

func loadConfigFile() ConfigFile {
  configDir, err := os.UserConfigDir()
  if err != nil {
    panic(err)
  }

  cfgPath := filepath.Join(configDir, "qj", "config.toml")

  file, err := os.Open(cfgPath)
  if err != nil {
    panic(err)
  }

  defer file.Close()

  content, err := ioutil.ReadAll(file)
  if err != nil {
    panic(err)
  }

  config := ConfigFile{}
  err = toml.Unmarshal(content, &config)
  if err != nil {
    panic(err)
  }

  return config
}

func validateConfigFile(cfg ConfigFile) error {
  msgs := make([]string, 0)

  if cfg.ApiKey == "" {
    msgs = append(msgs, "no API key was provided, you can create one here https://id.atlassian.com/manage/api-tokens")
  }
  if cfg.Email == "" {
    msgs = append(msgs, "no email was provided")
  }
  if cfg.BaseUrl == "" {
    msgs = append(msgs, "no base URL was provided")
  }

  if len(msgs) == 0 {
    return nil
  }

  var sb strings.Builder
  fmt.Fprintf(&sb, "errors were found in your config file:\n")
  for _, msg := range msgs {
    fmt.Fprintf(&sb, "  - %s\n", msg)
  }

  return errors.New(sb.String())
}

func LoadSavedQueries() (ConfigFile, error) {
  err := ensureConfigFile()
  if err != nil {
    return ConfigFile{}, err
  }

  config := loadConfigFile()
  err = validateConfigFile(config)
  if err != nil {
    return ConfigFile{}, err
  }

  return config, nil
}
