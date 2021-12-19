package comment

import "strings"

type CommentTree struct {
	Comment  string
	Children []CommentTree
}

type Model struct {
	Items []CommentTree
}

func (m Model) View() string {
	var result string
	for _, item := range m.Items {
		result += display_tree(item, 0) + "\n\n"
	}
	return result
}

func display_tree(ct CommentTree, depth int) string {

	var result string
	result += strings.Repeat("> ", depth) + ct.Comment + "\n\n"
	for _, item := range ct.Children {
		result += display_tree(item, depth+1) + "\n\n"
	}
	return result
}
