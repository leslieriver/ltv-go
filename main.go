package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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

type PaginationStatus struct {
	community string
	page      int
}

type model struct {
	state            int
	textInput        textinput.Model
	currentPost      viewport.Model
	paginationStatus PaginationStatus
	list             list.Model
	posts            []item
	lemmyapi         *lemmyapi.Client
}

func (m model) Init() tea.Cmd {
	return m.fetchPosts("")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering || (m.textInput.Focused() && msg.String() != "enter") {
			break
		}
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		} else if msg.String() == "enter" {
			if m.state == 0 {
				if m.textInput.Focused() {
					m.textInput.Blur()
					var trimmed = strings.Trim(m.textInput.Value(), "\n")
					m.paginationStatus.community = trimmed
					m.paginationStatus.page = 1
					return m, m.fetchPosts(trimmed)

				} else {
					m.textInput.Focus()
					m.textInput.SetValue("")
				}
			}
		} else if msg.String() == "ctrl+p" {
			if m.state == 0 {
				m.paginationStatus.page += 1
				return m, m.nextPage()

			}
		} else if msg.String() == "right" || msg.String() == "l" {
			if m.state == 0 && len(m.posts) > 0 {
				var selected = m.posts[m.list.Index()]
				str, err := glamour.Render(selected.body, "dark")
				if err != nil {
					return m, tea.Quit
				}
				if len(selected.body) > 0 {
					m.state = 1
				}
				m.currentPost.SetContent(str)
			}
			return m, nil

		} else if msg.String() == "left" || msg.String() == "h" {
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

		if !msg.Paginated {
			m.posts = []item{}
			m.list.SetItems([]list.Item{})
		}
		for _, val := range msg.Posts {
			m.posts = append(m.posts, item{title: fmt.Sprintf("%s", val.Post.Name), desc: val.Post.URL, body: val.Post.Body})
			m.list.InsertItem(len(m.list.Items()), m.posts[len(m.list.Items())])
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
	Err       error
	Posts     []lemmyapi.PostView
	Paginated bool
}

func (m model) fetchPosts(community string) tea.Cmd {
	return func() tea.Msg {
		p, err := m.lemmyapi.GetPosts(context.Background(), community, 1)
		if err != nil {
			return GotPosts{Err: err}
		}

		return GotPosts{Posts: p, Paginated: false}
	}
}
func (m model) nextPage() tea.Cmd {
	return func() tea.Msg {
		p, err := m.lemmyapi.GetPosts(context.Background(), m.paginationStatus.community, m.paginationStatus.page)
		if err != nil {
			return GotPosts{Err: err}
		}

		return GotPosts{Posts: p, Paginated: true}
	}
}

type listKeyMap struct {
	viewPost key.Binding
	goBack   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		viewPost: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "view post"),
		),
		goBack: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "back to list"),
		),
	}
}
func prependHttps(str string) string {
	if strings.HasPrefix(str, "https://") {
		return str
	} else {
		return "https://" + str
	}
}

func main() {
	var BaseUrl = prependHttps(os.Args[1])
	ti := textinput.NewModel()
	ti.Placeholder = "Community Name"
	ti.CharLimit = 156
	ti.Width = 20

	m := model{textInput: ti, list: list.NewModel([]list.Item{},
		list.NewDefaultDelegate(), 0, 0),
		lemmyapi: &lemmyapi.Client{HTTPClient: http.DefaultClient, BaseUrl: BaseUrl},
		posts:    []item{}, currentPost: viewport.Model{Width: 80, Height: 10},
		paginationStatus: PaginationStatus{community: "", page: 1}}
	m.list.Title = "Posts"
	var extraKeys = newListKeyMap()
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			extraKeys.goBack,
			extraKeys.viewPost,
		}
	}
	err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
