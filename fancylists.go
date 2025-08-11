package fancylists

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/brandenc40/romannumeral"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type listItemType int

const (
	notList listItemType = iota
	bulletList
	orderedList
	orderedListFancy
)

var skipListParserKey = parser.NewContextKey()
var emptyListItemWithBlankLines = parser.NewContextKey()
var listItemFlagValue interface{} = true

type FancyLists struct{}

func (e *FancyLists) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&fancyListParser{}, 100),   // Higher priority than default list parser (300)
		util.Prioritized(&fancyListItemParser{}, 101), // Higher priority than default list item parser (400)
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&fancyListHTMLRenderer{html.NewConfig()}, 500),
		util.Prioritized(&fancyListItemHTMLRenderer{html.NewConfig()}, 500),
	))
}

// parseListItem is based on Goldmark's parseListItem but extended to handle fancy lists
func parseListItem(line []byte) ([6]int, listItemType) {
	i := 0
	l := len(line)
	ret := [6]int{}
	for ; i < l && line[i] == ' '; i++ {
		c := line[i]
		if c == '\t' {
			return ret, notList
		}
	}
	if i > 3 {
		return ret, notList
	}
	ret[0] = 0
	ret[1] = i
	ret[2] = i
	var typ listItemType

	// Check for bullet list markers
	if i < l && (line[i] == '-' || line[i] == '*' || line[i] == '+') {
		i++
		ret[3] = i
		typ = bulletList
	} else if i < l {
		// Check for ordered list markers (numbers, letters, roman numerals, '#')
		start := i

		// Handle '#' as a special marker for continuing lists
		if line[i] == '#' {
			i++
			ret[3] = i
			if i < l && (line[i] == '.' || line[i] == ')') {
				i++
				ret[3] = i
			} else {
				return ret, notList
			}
			typ = orderedListFancy
		} else {
			// Check for numeric markers (1-9 digits)
			numStart := i
			for ; i < l && util.IsNumeric(line[i]); i++ {
			}
			if i > numStart && i-numStart <= 9 {
				// Found numeric marker
				ret[3] = i
				if i < l && (line[i] == '.' || line[i] == ')') {
					i++
					ret[3] = i
					typ = orderedList
				} else {
					return ret, notList
				}
			} else {
				// Check for alphabetic markers (letters only, 1-6 chars)
				i = start
				for ; i < l && i-start < 6 && unicode.IsLetter(rune(line[i])); i++ {
				}
				if i > start {
					// Found alphabetic marker
					ret[3] = i
					if i < l && (line[i] == '.' || line[i] == ')') {
						i++
						ret[3] = i
						typ = orderedListFancy
					} else {
						return ret, notList
					}
				} else {
					return ret, notList
				}
			}
		}
	} else {
		return ret, notList
	}

	if i < l && line[i] != '\n' {
		w, _ := util.IndentWidth(line[i:], 0)
		if w == 0 {
			return ret, notList
		}
	}
	if i >= l {
		ret[4] = -1
		ret[5] = -1
		return ret, typ
	}
	ret[4] = i
	ret[5] = len(line)
	if line[ret[5]-1] == '\n' && line[i] != '\n' {
		ret[5]--
	}
	return ret, typ
}

func matchesListItem(source []byte, strict bool) ([6]int, listItemType) {
	m, typ := parseListItem(source)
	if typ != notList && (!strict || strict && m[1] < 4) {
		return m, typ
	}
	return m, notList
}

func calcListOffset(source []byte, match [6]int) int {
	var offset int
	if match[4] < 0 || util.IsBlank(source[match[4]:]) { // list item starts with a blank line
		offset = 1
	} else {
		offset, _ = util.IndentWidth(source[match[4]:], match[4])
		if offset > 4 { // offseted codeblock
			offset = 1
		}
	}
	return offset
}

func lastOffset(node ast.Node) int {
	lastChild := node.LastChild()
	if lastChild != nil {
		return lastChild.(*ast.ListItem).Offset
	}
	return 0
}

// Helper functions for converting alphabetic and roman numeral markers to numbers
func alphabeticToNumber(s string) int {
	if len(s) == 0 {
		return 0
	}

	s = strings.ToLower(s)

	result := 0
	base := 26

	for i, char := range s {
		if char < 'a' || char > 'z' {
			return 0 // Invalid character
		}
		digit := int(char - 'a' + 1)
		if i == len(s)-1 {
			result += digit
		} else {
			result += digit * pow(base, len(s)-1-i)
		}
	}

	return result
}

func pow(base, exp int) int {
	result := 1
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}

func romanToNumber(s string) (int, bool) {
	// Check if it starts with valid roman numeral pattern
	if len(s) == 0 {
		return 0, false
	}

	first := strings.ToLower(s)[0]
	if first != 'i' {
		return 0, false // Only support roman numerals starting with 'i' (case insensitive)
	}

	// Convert to uppercase for parsing since romannumeral library expects uppercase
	upperS := strings.ToUpper(s)
	num, err := romannumeral.StringToInt(upperS)
	if err != nil {
		return 0, false
	}

	return num, true
}

type fancyListParser struct{}

func (b *fancyListParser) Trigger() []byte {
	// Include all possible list markers: bullets, numbers, letters, and hash
	triggers := []byte{'-', '+', '*', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '#'}

	// Add all letters
	for c := 'a'; c <= 'z'; c++ {
		triggers = append(triggers, byte(c))
	}
	for c := 'A'; c <= 'Z'; c++ {
		triggers = append(triggers, byte(c))
	}

	return triggers
}

func (b *fancyListParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	last := pc.LastOpenedBlock().Node
	if _, lok := last.(*ast.List); lok || pc.Get(skipListParserKey) != nil {
		pc.Set(skipListParserKey, nil)
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	match, typ := matchesListItem(line, true)
	if typ == notList {
		return nil, parser.NoChildren
	}

	start := -1
	var fltype *string

	if typ == orderedList {
		number := line[match[2] : match[3]-1]
		start, _ = strconv.Atoi(string(number))
	} else if typ == orderedListFancy {
		number := line[match[2] : match[3]-1]

		if string(number) == "#" {
			// For '#' marker, we'll determine type from context or default to numeric
			start = 1 // Default start
			// fltype remains nil for default behavior
		} else {
			// Check if it's a roman numeral first (must start with 'i' or 'I')
			if len(number) > 0 && (number[0] == 'i' || number[0] == 'I') {
				if romanNum, ok := romanToNumber(string(number)); ok {
					start = romanNum
					if unicode.IsLower(rune(number[0])) {
						fltype = &[]string{"i"}[0]
					} else {
						fltype = &[]string{"I"}[0]
					}
				} else {
					return nil, parser.NoChildren
				}
			} else if unicode.IsLetter(rune(number[0])) {
				// Alphabetic marker
				start = alphabeticToNumber(string(number))
				if start == 0 {
					return nil, parser.NoChildren
				}
				if unicode.IsLower(rune(number[0])) {
					fltype = &[]string{"a"}[0]
				} else {
					fltype = &[]string{"A"}[0]
				}
			}
		}
	}

	if ast.IsParagraph(last) && last.Parent() == parent {
		// we allow only lists starting with 1 to interrupt paragraphs.
		if typ == orderedList && start != 1 {
			return nil, parser.NoChildren
		}
		if typ == orderedListFancy && start != 1 {
			return nil, parser.NoChildren
		}
		//an empty list item cannot interrupt a paragraph:
		if match[4] < 0 || util.IsBlank(line[match[4]:match[5]]) {
			return nil, parser.NoChildren
		}
	}

	marker := line[match[3]-1]
	node := ast.NewList(marker)
	if start > -1 {
		node.Start = start
	}
	if fltype != nil {
		node.SetAttribute([]byte("type"), []byte(*fltype))
	}
	pc.Set(emptyListItemWithBlankLines, nil)
	return node, parser.HasChildren
}

func (b *fancyListParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	list := node.(*ast.List)
	line, _ := reader.PeekLine()
	if util.IsBlank(line) {
		if node.LastChild().ChildCount() == 0 {
			pc.Set(emptyListItemWithBlankLines, listItemFlagValue)
		}
		return parser.Continue | parser.HasChildren
	}

	offset := lastOffset(node)
	lastIsEmpty := node.LastChild().ChildCount() == 0
	indent, _ := util.IndentWidth(line, reader.LineOffset())

	if indent < offset || lastIsEmpty {
		if indent < 4 {
			match, typ := matchesListItem(line, false)
			if typ != notList && match[1]-offset < 4 {
				marker := line[match[3]-1]
				if !list.CanContinue(marker, typ == orderedList || typ == orderedListFancy) {
					return parser.Close
				}
				return parser.Continue | parser.HasChildren
			}
		}
		if !lastIsEmpty {
			return parser.Close
		}
	}

	if lastIsEmpty && indent < offset {
		return parser.Close
	}

	if pc.Get(emptyListItemWithBlankLines) != nil {
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

func (b *fancyListParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	list := node.(*ast.List)

	for c := node.FirstChild(); c != nil && list.IsTight; c = c.NextSibling() {
		if c.FirstChild() != nil && c.FirstChild() != c.LastChild() {
			for c1 := c.FirstChild().NextSibling(); c1 != nil; c1 = c1.NextSibling() {
				if c1.HasBlankPreviousLines() {
					list.IsTight = false
					break
				}
			}
		}
		if c != node.FirstChild() {
			if c.HasBlankPreviousLines() {
				list.IsTight = false
			}
		}
	}

	if list.IsTight {
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			for gc := child.FirstChild(); gc != nil; {
				paragraph, ok := gc.(*ast.Paragraph)
				gc = gc.NextSibling()
				if ok {
					textBlock := ast.NewTextBlock()
					textBlock.SetLines(paragraph.Lines())
					child.ReplaceChild(child, paragraph, textBlock)
				}
			}
		}
	}
}

func (b *fancyListParser) CanInterruptParagraph() bool {
	return true
}

func (b *fancyListParser) CanAcceptIndentedLine() bool {
	return false
}

type fancyListItemParser struct{}

func (b *fancyListItemParser) Trigger() []byte {
	// Include all possible list markers: bullets, numbers, letters, and hash
	triggers := []byte{'-', '+', '*', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '#'}

	// Add all letters
	for c := 'a'; c <= 'z'; c++ {
		triggers = append(triggers, byte(c))
	}
	for c := 'A'; c <= 'Z'; c++ {
		triggers = append(triggers, byte(c))
	}

	return triggers
}

func (b *fancyListItemParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	list, lok := parent.(*ast.List)
	if !lok { // list item must be a child of a list
		return nil, parser.NoChildren
	}
	offset := lastOffset(list)
	line, _ := reader.PeekLine()
	match, typ := matchesListItem(line, false)
	if typ == notList {
		return nil, parser.NoChildren
	}
	if match[1]-offset > 3 {
		return nil, parser.NoChildren
	}

	pc.Set(emptyListItemWithBlankLines, nil)

	itemOffset := calcListOffset(line, match)
	node := ast.NewListItem(match[3] + itemOffset)

	// Set the value attribute for fancy lists
	if typ == orderedList || typ == orderedListFancy {
		itemNumber := list.ChildCount() + list.Start
		node.SetAttribute([]byte("value"), []byte(strconv.Itoa(itemNumber)))
	}

	if match[4] < 0 || util.IsBlank(line[match[4]:match[5]]) {
		return node, parser.NoChildren
	}

	pos, padding := util.IndentPosition(line[match[4]:], match[4], itemOffset)
	child := match[3] + pos
	reader.AdvanceAndSetPadding(child, padding)
	return node, parser.HasChildren
}

func (b *fancyListItemParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if util.IsBlank(line) {
		reader.AdvanceToEOL()
		return parser.Continue | parser.HasChildren
	}

	offset := lastOffset(node.Parent())
	isEmpty := node.ChildCount() == 0 && pc.Get(emptyListItemWithBlankLines) != nil
	indent, _ := util.IndentWidth(line, reader.LineOffset())
	if (isEmpty || indent < offset) && indent < 4 {
		_, typ := matchesListItem(line, true)
		// new list item found
		if typ != notList {
			pc.Set(skipListParserKey, listItemFlagValue)
			return parser.Close
		}
		if !isEmpty {
			return parser.Close
		}
	}
	pos, padding := util.IndentPosition(line, reader.LineOffset(), offset)
	reader.AdvanceAndSetPadding(pos, padding)

	return parser.Continue | parser.HasChildren
}

func (b *fancyListItemParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (b *fancyListItemParser) CanInterruptParagraph() bool {
	return true
}

func (b *fancyListItemParser) CanAcceptIndentedLine() bool {
	return false
}

type fancyListHTMLRenderer struct {
	html.Config
}

func (r *fancyListHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindList, r.renderList)
}

func (r *fancyListHTMLRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString(tag)

		if n.IsOrdered() {
			// Add fancy class and attributes
			_, _ = w.WriteString(` class="fancy`)

			// Determine list type class
			if typeAttr, ok := n.AttributeString("type"); ok {
				typeBytes, ok := typeAttr.([]byte)
				if !ok {
					// Handle string case
					if typeStr, ok := typeAttr.(string); ok {
						typeBytes = []byte(typeStr)
					}
				}
				if typeBytes != nil {
					typeStr := string(typeBytes)
					switch typeStr {
					case "a":
						_, _ = w.WriteString(` fl-lcalpha"`)
					case "A":
						_, _ = w.WriteString(` fl-ucalpha"`)
					case "i":
						_, _ = w.WriteString(` fl-lcroman"`)
					case "I":
						_, _ = w.WriteString(` fl-ucroman"`)
					default:
						_, _ = w.WriteString(` fl-num"`)
					}
				} else {
					_, _ = w.WriteString(` fl-num"`)
				}
			} else {
				_, _ = w.WriteString(` fl-num"`)
			}

			if typeAttr, ok := n.AttributeString("type"); ok {
				_, _ = w.WriteString(` type="`)
				typeBytes, ok := typeAttr.([]byte)
				if !ok {
					// Handle string case
					if typeStr, ok := typeAttr.(string); ok {
						typeBytes = []byte(typeStr)
					}
				}
				if typeBytes != nil {
					_, _ = w.Write(typeBytes)
				}
				_ = w.WriteByte('"')
			} else {
				_, _ = w.WriteString(` type="1"`)
			}

			if n.Start != 1 {
				// Note: We don't add start attribute as per test expectations
				// The individual li elements have value attributes instead
			}
		}

		_, _ = w.WriteString(">\n")
	} else {
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(tag)
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

type fancyListItemHTMLRenderer struct {
	html.Config
}

func (r *fancyListItemHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindListItem, r.renderListItem)
}

func (r *fancyListItemHTMLRenderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<li")
		if valueAttr, ok := n.AttributeString("value"); ok {
			_, _ = w.WriteString(` value="`)
			valueBytes, ok := valueAttr.([]byte)
			if !ok {
				// Handle other types (string, int)
				if valueStr, ok := valueAttr.(string); ok {
					valueBytes = []byte(valueStr)
				} else if valueInt, ok := valueAttr.(int); ok {
					valueBytes = []byte(strconv.Itoa(valueInt))
				}
			}
			if valueBytes != nil {
				_, _ = w.Write(valueBytes)
			}
			_ = w.WriteByte('"')
		}
		_ = w.WriteByte('>')

		fc := n.FirstChild()
		if fc != nil {
			if _, ok := fc.(*ast.TextBlock); !ok {
				_ = w.WriteByte('\n')
			}
		}
	} else {
		_, _ = w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}