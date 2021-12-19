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
	"github.com/leslieriver/ltv-go/comment"
	"github.com/leslieriver/ltv-go/lemmyapi"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc, body string
	id                int
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
	comments         comment.Model
	currentComments  viewport.Model
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
			if m.state == 0 {
				if len(m.posts) > 0 {
					var selected = m.posts[m.list.Index()]
					var headString string = "# " + selected.title + "\n\n"
					str, err := glamour.Render(headString+selected.body, "dark")
					if err != nil {
						return m, tea.Quit
					}

					if len(selected.body) > 0 {
						m.state = 1
						m.currentPost.SetContent(str)
					} else {
						return m, m.fetchComments(m.posts[m.list.Index()].id)
					}

				}
			} else if m.state == 1 {
				return m, m.fetchComments(m.posts[m.list.Index()].id)
			}
			return m, nil

		} else if msg.String() == "left" || msg.String() == "h" {
			if m.state == 1 {
				m.state = 0
			} else if m.state == 2 {
				var selected = m.posts[m.list.Index()]
				if len(selected.body) > 0 {
					m.state = 1
				} else {
					m.state = 0
				}
			}
			return m, nil

		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.currentPost.Width = msg.Width - left - right
		m.currentPost.Height = msg.Height - top - bottom
		m.currentComments.Width = msg.Width - left - right
		m.currentComments.Height = msg.Height - top - bottom
	case GotPosts:
		if err := msg.Err; err != nil {
			return m, nil
		}

		if !msg.Paginated {
			m.posts = []item{}
			m.list.SetItems([]list.Item{})
		}
		for _, val := range msg.Posts {
			m.posts = append(m.posts, item{title: fmt.Sprintf("%s", val.Post.Name), desc: val.Post.URL, body: val.Post.Body, id: val.Post.ID})
			m.list.InsertItem(len(m.list.Items()), m.posts[len(m.list.Items())])
		}
		return m, nil

	case GotComments:
		if err := msg.Err; err != nil {
			return m, nil
		}
		m.comments.Items = msg.Comments
		var headString string = "# Comments\n\n"
		if len(m.comments.Items) == 0 {
			headString = "# No comments\n\n"
		}
		var str, err = glamour.Render(headString+m.comments.View(), "dark")
		if err != nil {
			return m, tea.Quit
		}

		m.currentComments.SetContent(str)
		m.state = 2

		return m, nil
	}

	var cmdlist, cmdtext, cmdpost, cmdcomments tea.Cmd
	if m.textInput.Focused() {
		m.textInput, cmdtext = m.textInput.Update(msg)
		return m, cmdtext
	} else {
		if m.state == 1 {
			m.currentPost, cmdpost = m.currentPost.Update(msg)
			return m, cmdpost
		} else if m.state == 2 {
			m.currentComments, cmdcomments = m.currentComments.Update(msg)
			return m, cmdcomments
		} else {
			m.list, cmdlist = m.list.Update(msg)
			return m, cmdlist
		}
	}
}

func (m model) View() string {
	if m.state == 1 {
		return docStyle.Render(m.currentPost.View())
	} else if m.state == 2 {
		return docStyle.Render(m.currentComments.View())
	} else {
		return docStyle.Render(m.textInput.View() + "\n\n" + m.list.View())
	}
}

type GotPosts struct {
	Err       error
	Posts     []lemmyapi.PostView
	Paginated bool
}
type GotComments struct {
	Err      error
	Comments []comment.CommentTree
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
func (m model) fetchComments(id int) tea.Cmd {
	return func() tea.Msg {
		c, err := m.lemmyapi.GetComments(context.Background(), id)
		if err != nil {
			return GotComments{Err: err}
		}
		var ct = build_tree(c)
		return GotComments{Comments: ct}
	}
}
func build_tree(c []lemmyapi.CommentView) []comment.CommentTree {
	var c_map = make(map[int]comment.CommentTree)
	for _, item := range c {
		c_map[item.Comment.ID] = comment.CommentTree{Comment: item.Comment.Content, Children: []comment.CommentTree{}}
	}
	var tree = []comment.CommentTree{}
	for _, item := range c {
		var child = c_map[item.Comment.ID]
		var parent_id = item.Comment.ParentID
		if parent_id != nil {
			var parent = c_map[*parent_id]
			parent.Children = append(parent.Children, child)
			c_map[*parent_id] = parent
		} else {
			tree = append(tree, child)
		}
	}

	return tree
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
	nextScreen     key.Binding
	prevScreen     key.Binding
	communityInput key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		nextScreen: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "next screen"),
		),
		prevScreen: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "prev screen"),
		),
		communityInput: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↵", "open community input"),
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
	var BaseUrl = "https://lemmy.ml"
	if len(os.Args) > 1 {
		BaseUrl = prependHttps(os.Args[1])
	}
	ti := textinput.NewModel()
	ti.Placeholder = "Community Name"
	ti.CharLimit = 156
	ti.Width = 20

	m := model{textInput: ti, list: list.NewModel([]list.Item{},
		list.NewDefaultDelegate(), 0, 0),
		lemmyapi: &lemmyapi.Client{HTTPClient: http.DefaultClient, BaseUrl: BaseUrl},
		posts:    []item{}, currentPost: viewport.Model{Width: 80, Height: 10},
		currentComments:  viewport.Model{Width: 80, Height: 10},
		paginationStatus: PaginationStatus{community: "", page: 1}}
	m.list.Title = "Posts"
	var extraKeys = newListKeyMap()
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			extraKeys.nextScreen,
			extraKeys.prevScreen,
			extraKeys.communityInput,
		}
	}
	err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
