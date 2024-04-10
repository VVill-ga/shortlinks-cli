# Shortlinks CLI

Shortlinks CLI is a cross-platform terminal based front-end to the
[Shortlinks API](https://github.com/VVill-ga/shortlinks) written in Go.
Shortlinks CLI uses the [Go-Arg](https://github.com/alexflint/go-arg)
library to help manage command line arguments. Usage is shown below:

```bash
shorten [-s shortlinks_url] [-c requested_shortcode] url1 url2...

# Shows help info
shorten -h
shorten --help

# Sets default shortlink server url
shorten -s shortlinks_url
shorten --set-server shortlinks_url

# Uses provided shortlink server url for this shortening only
shorten -s shortlinks_url url1 url2 ...
shorten --set-server shortlinks_url url1 url2 ...

# Creates shortlink to provided url
shorten url1 url2 ...

# Requsts `shortcode` to redirect to `url`
# NOTE: Can only be used with one url
shorten -c requested_shortcode url
shorten --request-code requested_shortcode url
```

The one bit of configuration used by this program (the default server url)
is stored in plaintext in a hidden file `.shortlinks_server` in the users home
directory (found using Go's `os.UserHomeDir()`).