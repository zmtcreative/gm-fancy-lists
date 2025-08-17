// Package fancylists provides a Goldmark extension for Pandoc-style "fancy lists".
// Supports alphabetic (a., A.), roman numeral (i., I.), and hash continuation (#.) markers.
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

// listItemType represents the type of list item marker detected during parsing.
type listItemType int

// List item type constants for different marker styles.
const (
	notList listItemType = iota
	bulletList
	orderedList
	orderedListFancy
)

// Internal parser context keys for state management.
var (
	skipListParserKey           = parser.NewContextKey()
	emptyListItemWithBlankLines = parser.NewContextKey()
	listItemFlagValue           interface{} = true
)

// FancyLists extends Goldmark to support fancy list markers.
type FancyLists struct{}

// Extend implements goldmark.Extender interface to register parsers and renderers.
func (e *FancyLists) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&fancyListParser{}, 100),     // Higher priority than default list parser (300)
		util.Prioritized(&fancyListItemParser{}, 101), // Higher priority than default list item parser (400)
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&fancyListHTMLRenderer{html.NewConfig()}, 500),
		util.Prioritized(&fancyListItemHTMLRenderer{html.NewConfig()}, 500),
	))
}

// parseListItem analyzes a line of text to determine if it contains a list item marker.
// Returns position information and list item type.
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

func getListTypeFromMarker(markerBytes []byte, typ listItemType) (string, string) {
	marker := string(markerBytes)

	if typ == orderedList {
		return "1", "fl-num"
	}

	if typ == orderedListFancy {
		if marker == "#" {
			// For '#' marker, we default to numeric unless context suggests otherwise
			return "1", "fl-num"
		} else if len(marker) > 0 {
			// Check if it's a roman numeral first (must start with 'i' or 'I')
			if marker[0] == 'i' || marker[0] == 'I' {
				if _, ok := romanToNumber(marker); ok {
					if unicode.IsLower(rune(marker[0])) {
						return "i", "fl-lcroman"
					} else {
						return "I", "fl-ucroman"
					}
				}
			}
			// Otherwise it's alphabetic
			if unicode.IsLower(rune(marker[0])) {
				return "a", "fl-lcalpha"
			} else {
				return "A", "fl-ucalpha"
			}
		}
	}

	// Default fallback
	return "1", "fl-num"
}

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

	// Only support roman numerals starting with 'i' (case insensitive)
	// This means: i, ii, iii, iv (lowercase) or I, II, III, IV (uppercase)
	// But NOT: vi, vii, etc. (those are treated as alphabetic)
	first := strings.ToLower(s)[0]
	if first != 'i' {
		return 0, false
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

	switch typ {
	case orderedList:
		number := line[match[2] : match[3]-1]
		start, _ = strconv.Atoi(string(number))
	case orderedListFancy:
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
		// we allow only lists starting with 1 to interrupt paragraphs,
		// but this restriction doesn't apply to nested lists (inside list items)
		if _, isListItem := parent.(*ast.ListItem); !isListItem {
			if typ == orderedList && start != 1 {
				return nil, parser.NoChildren
			}
			if typ == orderedListFancy && start != 1 {
				return nil, parser.NoChildren
			}
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

				// Check if the list can continue with this marker type
				if !list.CanContinue(marker, typ == orderedList || typ == orderedListFancy) {
					return parser.Close
				}

				// For ordered lists, check if the type has changed
				if typ == orderedList || typ == orderedListFancy {
					markerBytes := line[match[2] : match[3]-1]
					markerStr := string(markerBytes)

					// If it's a '#' marker, it should continue the current list type
					if markerStr != "#" {
						// Get current list type
						currentType := "1" // default
						if currentTypeAttr, ok := list.AttributeString("type"); ok {
							if typeBytes, ok := currentTypeAttr.([]byte); ok {
								currentType = string(typeBytes)
							} else if typeStr, ok := currentTypeAttr.(string); ok {
								currentType = typeStr
							}
						}

						// For specific markers (non-#), determine expected type with context awareness
						var expectedType string

						// Handle the ambiguous case of 'i'/'I'
						if len(markerStr) == 1 && (markerStr == "i" || markerStr == "I") {
							// If current list is alphabetic AND same case, treat 'i'/'I' as alphabetic
							// If current list is different case alphabetic, numeric, or roman, treat 'i'/'I' as roman
							if (currentType == "a" && markerStr == "i") || (currentType == "A" && markerStr == "I") {
								// Same case alphabetic - continue as alphabetic
								expectedType = currentType
							} else {
								// Different case, numeric, or roman - treat as roman numeral
								if markerStr == "i" {
									expectedType = "i"
								} else {
									expectedType = "I"
								}
							}
						} else {
							// For non-ambiguous cases, use normal logic
							expectedType, _ = getListTypeFromMarker(markerBytes, typ)
						}

						// If types don't match, close this list to start a new one
						if expectedType != currentType {
							return parser.Close
						}
					}
					// If it's '#', continue with current list type (no type change)
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

// fancyListHTMLRenderer provides HTML rendering for fancy lists.
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

		// Handle class attribute - combine fancy list classes with user-defined classes
		var classValues []string

		if n.IsOrdered() {
			// Add fancy class and determine list type class
			classValues = append(classValues, "fancy")

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
						classValues = append(classValues, "fl-lcalpha")
					case "A":
						classValues = append(classValues, "fl-ucalpha")
					case "i":
						classValues = append(classValues, "fl-lcroman")
					case "I":
						classValues = append(classValues, "fl-ucroman")
					default:
						classValues = append(classValues, "fl-num")
					}
				} else {
					classValues = append(classValues, "fl-num")
				}
			} else {
				classValues = append(classValues, "fl-num")
			}
		}

		// Add user-defined class attributes from goldmark-attributes extension
		if classAttr, ok := n.AttributeString("class"); ok {
			if classBytes, ok := classAttr.([]byte); ok {
				classValues = append(classValues, string(classBytes))
			} else if classStr, ok := classAttr.(string); ok {
				classValues = append(classValues, classStr)
			}
		}

		// Write the class attribute if we have any classes
		if len(classValues) > 0 {
			_, _ = w.WriteString(` class="`)
			for i, class := range classValues {
				if i > 0 {
					_ = w.WriteByte(' ')
				}
				_, _ = w.WriteString(class)
			}
			_ = w.WriteByte('"')
		}

		// Handle ordered list specific attributes
		if n.IsOrdered() {
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
				// Add start attribute to the ol element
				_, _ = w.WriteString(` start="`)
				_, _ = w.WriteString(strconv.Itoa(n.Start))
				_ = w.WriteByte('"')
			} else {
				// Always add start="1" for consistency
				_, _ = w.WriteString(` start="1"`)
			}
		}

		// Handle all other attributes from goldmark-attributes extension
		if n.Attributes() != nil {
			for _, attr := range n.Attributes() {
				name := string(attr.Name)
				// Skip attributes we've already handled
				if name != "class" && name != "type" {
					_, _ = w.WriteString(` `)
					_, _ = w.WriteString(name)
					_, _ = w.WriteString(`="`)
					// Handle different value types
					if valueBytes, ok := attr.Value.([]byte); ok {
						_, _ = w.Write(valueBytes)
					} else if valueStr, ok := attr.Value.(string); ok {
						_, _ = w.WriteString(valueStr)
					}
					_ = w.WriteByte('"')
				}
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

// fancyListItemHTMLRenderer provides HTML rendering for fancy list items.
type fancyListItemHTMLRenderer struct {
	html.Config
}

func (r *fancyListItemHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindListItem, r.renderListItem)
}

func (r *fancyListItemHTMLRenderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<li")
		// No value attribute - the start attribute on the parent ol handles numbering
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