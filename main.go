package main

import (
  "context"
  "flag"
  "fmt"
  "log"
  "os"
  "os/exec"
  "time"
  "strings"
  "golang.org/x/sync/errgroup"
  "github.com/alligator/qj/src"
)

type FetchResult struct {
  Issues []qj.JiraIssue
  Query  qj.Query
}

func fetchSavedQuery(out chan FetchResult, ctx qj.JiraApiContext, query qj.Query) error {
  issues, err := qj.FetchIssuesJql(ctx, query.Jql)
  if err != nil {
    return err
  }
  out <- FetchResult{issues, query}
  return nil
}

func showSpinner() chan bool {
  spinnerFrames := []string{
    "[=   ]",
    "[==  ]",
    "[ == ]",
    "[  ==]",
    "[   =]",
    "[  ==]",
    "[ == ]",
    "[==  ]",
  }
  ticker := time.NewTicker(time.Millisecond * 100)
  done := make(chan bool)

  go func() {
    idx := 0
    for {
      select {
      case <-done:
        return
      case <-ticker.C:
        idx = (idx + 1) % len(spinnerFrames)
        fmt.Printf("\r%s", spinnerFrames[idx])
      }
    }
  }()

  return done
}

func main() {
  // remove the date from the logs
  log.SetFlags(0)

  configHelp := flag.Bool("c", false, "show the location of the config file")
  editConfig := flag.Bool("e", false, "open the config file in an editor")
  adHocQuery := flag.String("q", "", "run an ad-hoc jql query")

  flag.Parse()

  if *configHelp {
    fmt.Printf("config file location: %s\n", qj.ConfigPath())
    os.Exit(0)
  }

  if *editConfig {
    editor, found := os.LookupEnv("EDITOR")

    if !found {
      fmt.Println("no $EDITOR set, trying nvim")
      editor = "nvim"
    }

    cmd := exec.Command(editor, qj.ConfigPath())
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
      log.Fatal(err)
    }
    os.Exit(0)
  }

  config, err := qj.LoadSavedQueries()

  if err != nil {
    log.Fatal(err)
  }

  ctx := qj.JiraApiContext{
    Email:   config.Email,
    ApiKey:  config.ApiKey,
    BaseUrl: config.BaseUrl,
  }

  queries := config.Queries
  if *adHocQuery != "" {
    queries = []qj.Query{
      qj.Query{
        Name: "Command Line Query",
        Jql:  *adHocQuery,
      },
    }
  }

  spinnerChan := showSpinner()

  ch := make(chan FetchResult, len(queries))
  errs, _ := errgroup.WithContext(context.Background())
  for _, query := range queries {
    // https://golang.org/doc/faq#closures_and_goroutines
    query := query
    errs.Go(func() error {
      return fetchSavedQuery(ch, ctx, query)
    })
  }

  if err := errs.Wait(); err != nil {
    log.Fatal(err)
  }

  // collect the results
  results := make(map[string][]qj.JiraIssue)
  for i := 0; i < len(queries); i++ {
    result := <-ch
    results[result.Query.Name] = result.Issues
  }

  spinnerChan <- true

  // clean up the spinner
  fmt.Print("\r")

  // gather up all of the output columns
  columns := make(map[string][][]string)
  for queryName, issues := range results {
    cells := make([][]string, 0)
    for _, issue := range issues {
      cells = append(cells, []string {
        issue.Key,
        issue.Fields.Summary,
        issue.Fields.Assignee.DisplayName,
        strings.Join(issue.Fields.Labels, " "),
      })
    }
    columns[queryName] = cells
  }

  // figure out the max width of each column
  columnWidths := make([]int, 0)
  for _, rows := range columns {
    for _, cells := range rows {
      for idx, cell := range cells {
        if len(columnWidths) <= idx {
          columnWidths = append(columnWidths, len(cell))
        }

        if len(cell) > columnWidths[idx] {
          columnWidths[idx] = len(cell)
        }
      }
    }
  }

  // print the stuff
  for _, query := range queries {
    fmt.Printf("\x1b[93m%s\x1b[0m\n", query.Name)
    for _, cells := range columns[query.Name] {
      fmt.Print("  ")
      for idx, cell := range cells {
        formatString := fmt.Sprintf("%%-%ds  ", columnWidths[idx])
        fmt.Printf(formatString, cell)
      }
      fmt.Println()
    }
    fmt.Println()
  }
}
