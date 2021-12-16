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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leslieriver/ltv-go/lemmyapi"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	textInput textinput.Model
	list      list.Model
	posts     []list.Item
	lemmyapi  *lemmyapi.Client
}

func (m model) Init() tea.Cmd {
	return m.fetchPosts("")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			tea.Quit()
		} else if msg.String() == "enter" {
			if m.textInput.Focused() {
				m.textInput.Blur()
				var trimmed = strings.Trim(m.textInput.Value(), "\n")
				return m, m.fetchPosts(trimmed)

			} else {
				m.textInput.Focus()
				m.textInput.SetValue("")
			}
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	case GotPosts:
		if err := msg.Err; err != nil {
			return m, nil
		}

		m.posts = []list.Item{}
		m.list.SetItems([]list.Item{})
		for idx, val := range msg.Posts {
			m.posts = append(m.posts, item{title: fmt.Sprintf("[%s] %s", strconv.Itoa(idx+1), val.Post.Name), desc: val.Post.URL})
			m.list.InsertItem(idx, m.posts[idx])
		}
		return m, nil
	}

	var cmdlist, cmdtext tea.Cmd
	if m.textInput.Focused() {
		m.textInput, cmdtext = m.textInput.Update(msg)
	} else {
		m.list, cmdlist = m.list.Update(msg)
	}
	return m, tea.Batch(cmdlist, cmdtext)
}

func (m model) View() string {
	return docStyle.Render(m.textInput.View() + "\n\n" + m.list.View())
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

	m := model{textInput: ti, list: list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0), lemmyapi: &lemmyapi.Client{HTTPClient: http.DefaultClient, BaseUrl: "https://fapsi.be"}, posts: []list.Item{}}
	m.list.Title = "Posts"
	err := tea.NewProgram(m, tea.WithAltScreen()).Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
