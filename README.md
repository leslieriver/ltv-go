# ltv-go
my attempt at rewriting [ltv](https://github.com/lunatichacker/lemmy-terminal-viewer) in go.

## Install and Usage

Download binaries for your os from [releases](https://github.com/leslieriver/ltv-go/releases/) or build yourself if you have go installed

once installed run
``` ltv-go [instance-name]```

### Navigation

ltv-go uses vim like keybindings and arrows keys to navigate between a list of posts, post-body and comments
everything is explained in the help text within the app itself but i'll list some keys that might be unintuitive

* Load more posts: ctrl+p
* Show a text prompt (for fetching posts from a community): enter
* Submit your text input : enter


### Lacking Feautures (compared to ltv)
* Auth
* Configs
* Theming

### Additonal Features (compared to ltv)
* Pagination
* Markdown Support


### Built With

ltv-go is made possible only because of [BubbbleTea](https://github.com/charmbracelet/bubbletea)


