package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/leslieriver/ltv-go/lemmyapi"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc, body string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	state       int
	textInput   textinput.Model
	currentPost viewport.Model
	list        list.Model
	posts       []item
	lemmyapi    *lemmyapi.Client
}

func (m model) Init() tea.Cmd {
	return m.fetchPosts("")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "enter" {
			if m.state == 0 {
				if m.textInput.Focused() {
					m.textInput.Blur()
					var trimmed = strings.Trim(m.textInput.Value(), "\n")
					return m, m.fetchPosts(trimmed)

				} else {
					m.textInput.Focus()
					m.textInput.SetValue("")
				}
			}
		} else if msg.String() == "right" {
			var selected = m.posts[m.list.Cursor()]
			str, err := glamour.Render(selected.body, "dark")
			if err != nil {
				return m, tea.Quit
			}
			if len(selected.body) > 0 {
				m.state = 1
			}
			m.currentPost.SetContent(str)
			return m, nil

		} else if msg.String() == "left" {
			m.state = 0
			return m, nil

		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.currentPost.Width = msg.Width - left - right
		m.currentPost.Height = msg.Height - top - bottom
	case GotPosts:
		if err := msg.Err; err != nil {
			return m, nil
		}

		m.posts = []item{}
		m.list.SetItems([]list.Item{})
		for idx, val := range msg.Posts {
			m.posts = append(m.posts, item{title: fmt.Sprintf("[%s] %s", strconv.Itoa(idx+1), val.Post.Name), desc: val.Post.URL, body: val.Post.Body})
			m.list.InsertItem(idx, m.posts[idx])
		}
		return m, nil
	}

	var cmdlist, cmdtext, cmdpost tea.Cmd
	if m.textInput.Focused() {
		m.textInput, cmdtext = m.textInput.Update(msg)
		return m, cmdtext
	} else {
		if m.state == 1 {
			m.currentPost, cmdpost = m.currentPost.Update(msg)
			return m, cmdpost
		} else {
			m.list, cmdlist = m.list.Update(msg)
			return m, cmdlist
		}
	}
}

func (m model) View() string {
	if m.state == 1 {
		return docStyle.Render(m.currentPost.View())
	} else {
		return docStyle.Render(m.textInput.View() + "\n\n" + m.list.View())
	}
}

type GotPosts struct {
	Err   error
	Posts []lemmyapi.PostView
}

func (m model) fetchPosts(community string) tea.Cmd {
	return func() tea.Msg {
		p, err := m.lemmyapi.GetPosts(context.Background(), community)
		if err != nil {
			return GotPosts{Err: err}
		}

		return GotPosts{Posts: p}
	}
}

func main() {
	ti := textinput.NewModel()
	ti.Placeholder = "Community Name"
	ti.CharLimit = 156
	ti.Width = 20

	m := model{textInput: ti, list: list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0), lemmyapi: &lemmyapi.Client{HTTPClient: http.DefaultClient, BaseUrl: "https://fapsi.be"}, posts: []item{}, currentPost: viewport.Model{Width: 80, Height: 10}}
	m.list.Title = "Posts"
	m.list.DisableQuitKeybindings()
	err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
