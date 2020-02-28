# qj - query jira

qj runs JQL queries against JIRA in the terminal. You can run a set of queries saved in a config file (the default behaviour) or an ad-hoc query.

## usage
```
Usage of qj:
  -c    show the location of the config file
  -e    open the config file in an editor
  -q string
        run an ad-hoc jql query
```

The first time qj runs, it will generate a TOML config file and print the location. The file looks like this:

```toml
Email = "your JIRA account email address"
ApiKey = "your JIRA API key"
BaseUrl = "the URL to your JIRA instance"

[[Queries]]
  Name = "Review"
  Jql = "sprint in openSprints() and status = Review"

[[Queries]]
  Name = "Testing"
  Jql = "sprint in openSprints() and status = Testing"
```

The format is quite self explanatory. The `[[Queries]]` section can be repeated to add more queries.
