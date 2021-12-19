# ltv-go
my attempt at rewriting [ltv](https://github.com/lunatichacker/lemmy-terminal-viewer) in go.
<p float=left>
<img src=https://user-images.githubusercontent.com/94739408/146679824-0a530acc-ee74-45be-b782-10677964e559.png width =300 />
<img src =https://user-images.githubusercontent.com/94739408/146679787-52c3980f-99d2-480b-a81e-5b3bd8897f4a.png  width =200/>
 <img src =https://user-images.githubusercontent.com/94739408/146679748-4014674a-8d3e-4c53-8a53-c99b8c07eccb.png width =200 />
</p>

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


