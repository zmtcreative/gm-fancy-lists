// Package fancylists provides a Goldmark extension that adds support for Pandoc-style "fancy lists".
//
// This extension extends Goldmark's list parsing to support additional list marker types
// including alphabetic markers (a., b., c. or A., B., C.), roman numeral markers
// (i., ii., iii. or I., II., III.), and hash continuation markers (#.) that continue
// the current list numbering sequence.
//
// # Features
//
//   - Numeric lists: 1., 2., 3., etc.
//   - Lowercase alphabetic lists: a., b., c., etc.
//   - Uppercase alphabetic lists: A., B., C., etc.
//   - Lowercase roman numeral lists: i., ii., iii., etc. (starting with 'i' only)
//   - Uppercase roman numeral lists: I., II., III., etc. (starting with 'I' only)
//   - Hash continuation: #. continues the current list type and numbering
//   - Automatic type detection and list separation when types change
//   - Proper HTML rendering with CSS classes for styling
//
// # Usage
//
//	import (
//		"github.com/yuin/goldmark"
//		"github.com/ZMT-Creative/gm-fancy-lists"
//	)
//
//	md := goldmark.New(
//		goldmark.WithExtensions(
//			&fancylists.FancyLists{},
//		),
//	)
//
// # HTML Output
//
// The extension generates HTML with specific CSS classes for each list type:
//   - fl-num: Numeric lists
//   - fl-lcalpha: Lowercase alphabetic lists
//   - fl-ucalpha: Uppercase alphabetic lists
//   - fl-lcroman: Lowercase roman numeral lists
//   - fl-ucroman: Uppercase roman numeral lists
//
// All ordered lists include a "fancy" class and appropriate type and start attributes.
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
	// notList indicates the line is not a list item
	notList listItemType = iota
	// bulletList indicates an unordered list item (-, *, +)
	bulletList
	// orderedList indicates a standard numeric list item (1., 2., 3.)
	orderedList
	// orderedListFancy indicates an extended list item (a., A., i., I., #.)
	orderedListFancy
)

// Internal parser context keys for state management.
var (
	// skipListParserKey is used to prevent recursive list parsing
	skipListParserKey = parser.NewContextKey()
	// emptyListItemWithBlankLines tracks list items that start with blank lines
	emptyListItemWithBlankLines = parser.NewContextKey()
	// listItemFlagValue is a sentinel value used for context flags
	listItemFlagValue interface{} = true
)

// FancyLists is the main extension struct that implements goldmark.Extender.
//
// Add this extension to a Goldmark instance to enable fancy list support:
//
//	md := goldmark.New(
//		goldmark.WithExtensions(&FancyLists{}),
//	)
type FancyLists struct{}

// Extend implements goldmark.Extender interface.
// It registers the fancy list parsers and renderers with the Goldmark instance.
//
// The extension registers:
//   - fancyListParser: Handles list container parsing with priority 100
//   - fancyListItemParser: Handles individual list item parsing with priority 101
//   - fancyListHTMLRenderer: Renders list containers to HTML with CSS classes
//   - fancyListItemHTMLRenderer: Renders list items to HTML
//
// These parsers have higher priority than Goldmark's default list parsers
// to ensure fancy list syntax is processed correctly.
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

// parseListItem analyzes a line of text to determine if it contains a list item marker
// and extracts relevant information about the marker and content.
//
// This function is based on Goldmark's parseListItem but extended to handle fancy list types
// including alphabetic markers, roman numerals, and hash continuation markers.
//
// Returns:
//   - [6]int: Array containing position information [0, indentWidth, markerStart, markerEnd, contentStart, contentEnd]
//   - listItemType: The type of list item detected (notList, bulletList, orderedList, orderedListFancy)
//
// The function handles:
//   - Bullet list markers: -, *, +
//   - Numeric markers: 1-9 digits followed by . or )
//   - Alphabetic markers: 1-6 letters followed by . or )
//   - Hash continuation markers: # followed by . or )
//   - Proper indentation validation (max 3 spaces)
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

// matchesListItem checks if a line contains a valid list item marker.
// It wraps parseListItem with additional validation based on strict mode.
//
// Parameters:
//   - source: The line of text to analyze
//   - strict: If true, applies stricter indentation rules (max 3 spaces)
//
// Returns the same values as parseListItem if valid, or notList type if invalid.
func matchesListItem(source []byte, strict bool) ([6]int, listItemType) {
	m, typ := parseListItem(source)
	if typ != notList && (!strict || strict && m[1] < 4) {
		return m, typ
	}
	return m, notList
}

// calcListOffset calculates the content offset for a list item based on its marker.
// This determines how much subsequent lines need to be indented to be considered
// part of the same list item.
//
// Parameters:
//   - source: The source line containing the list item
//   - match: Position array from parseListItem
//
// Returns the calculated offset for content alignment.
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

// lastOffset returns the offset of the last list item in a list node.
// This is used to determine the expected indentation for continuing list items.
func lastOffset(node ast.Node) int {
	lastChild := node.LastChild()
	if lastChild != nil {
		return lastChild.(*ast.ListItem).Offset
	}
	return 0
}

// Helper functions for converting alphabetic and roman numeral markers to numbers

// getListTypeFromMarker determines the HTML type attribute and CSS class
// that should be used for a list based on its marker.
//
// This function analyzes the marker bytes and list item type to determine:
//   - The HTML type attribute value ("1", "a", "A", "i", "I")
//   - The corresponding CSS class suffix ("fl-num", "fl-lcalpha", etc.)
//
// Parameters:
//   - markerBytes: The raw marker text (e.g., "a", "i", "III", "#")
//   - typ: The list item type (orderedList or orderedListFancy)
//
// Returns:
//   - string: HTML type attribute value
//   - string: CSS class suffix for styling
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

// alphabeticToNumber converts alphabetic markers to their numeric position.
//
// This function converts alphabetic sequences like "a", "b", "z", "aa", "ab" etc.
// to their corresponding numeric positions (1, 2, 26, 27, 28, etc.).
// It supports both single and multi-character alphabetic markers.
//
// Parameters:
//   - s: The alphabetic marker string (e.g., "a", "z", "aa")
//
// Returns the numeric position, or 0 if the input is invalid.
//
// Examples:
//   - "a" -> 1
//   - "z" -> 26
//   - "aa" -> 27
//   - "ab" -> 28
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

// pow computes base raised to the power of exp using integer arithmetic.
// This is a helper function for alphabeticToNumber calculations.
func pow(base, exp int) int {
	result := 1
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}

// romanToNumber converts roman numeral strings to their numeric values.
//
// This function specifically handles roman numerals that begin with 'i' or 'I'
// (case insensitive) to distinguish them from alphabetic markers. Roman numerals
// that don't start with 'i'/'I' (like "vi", "VII") are treated as alphabetic markers.
//
// This design choice ensures that sequences like "vi. vii. viii." are treated as
// alphabetic rather than roman, following Pandoc's behavior.
//
// Parameters:
//   - s: The roman numeral string (e.g., "i", "ii", "III", "IV")
//
// Returns:
//   - int: The numeric value of the roman numeral
//   - bool: True if the string is a valid roman numeral starting with i/I
//
// Examples:
//   - "i" -> (1, true)
//   - "IV" -> (4, true)
//   - "vi" -> (0, false) - treated as alphabetic
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

// fancyListParser implements parser.BlockParser for parsing fancy list containers.
//
// This parser handles the creation and continuation of list containers that support
// extended marker types. It works in conjunction with fancyListItemParser to provide
// complete fancy list support.
//
// The parser:
//   - Detects when a new list should start
//   - Determines the list type based on the first marker
//   - Handles type changes that require closing the current list
//   - Manages nested list contexts
type fancyListParser struct{}

// Trigger returns the byte values that can trigger this parser.
// This includes all possible list marker characters: bullets, digits, letters, and hash.
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

// Open attempts to start a new list when a list marker is encountered.
//
// This method:
//   - Validates that a new list should start (not continuing an existing one)
//   - Analyzes the first marker to determine list type and starting value
//   - Handles special cases like paragraph interruption rules
//   - Creates and configures the new list node with appropriate attributes
//
// Parameters:
//   - parent: The parent AST node where the list would be added
//   - reader: Text reader for accessing the source content
//   - pc: Parser context for state management
//
// Returns:
//   - ast.Node: The new list node, or nil if no list should start
//   - parser.State: Parser state indicating whether the node has children
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

// Continue determines whether the current list should continue or close
// when encountering the next line.
//
// This method implements the core logic for fancy list type detection and
// automatic list separation. When the list type changes (e.g., from numeric
// to alphabetic), it closes the current list to allow a new one to start.
//
// Key features:
//   - Detects type changes between list items at the same level
//   - Handles context-aware disambiguation (e.g., 'i' as alphabetic vs roman)
//   - Preserves type continuity for hash (#) markers
//   - Manages proper indentation and nesting rules
//
// Parameters:
//   - node: The current list node being processed
//   - reader: Text reader for accessing the source content
//   - pc: Parser context for state management
//
// Returns parser.State indicating whether to continue, close, or handle children.
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

// Close finalizes the list node after all items have been processed.
//
// This method handles post-processing tasks such as:
//   - Determining if the list should be "tight" (no paragraph wrapping)
//   - Converting paragraphs to text blocks in tight lists
//   - Cleaning up the final list structure
//
// Parameters:
//   - node: The list node being closed
//   - reader: Text reader (unused in this implementation)
//   - pc: Parser context (unused in this implementation)
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

// CanInterruptParagraph returns true if this parser can interrupt a paragraph.
// Fancy lists can interrupt paragraphs under certain conditions.
func (b *fancyListParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine returns false as list parsers don't handle indented lines directly.
// Indented content is handled by the list item parser.
func (b *fancyListParser) CanAcceptIndentedLine() bool {
	return false
}

// fancyListItemParser implements parser.BlockParser for parsing individual list items.
//
// This parser handles the creation and continuation of list items within fancy lists.
// It works together with fancyListParser to provide complete list item processing
// including content parsing and proper nesting.
type fancyListItemParser struct{}

// Trigger returns the byte values that can trigger this parser.
// Same as fancyListParser since list items use the same marker characters.
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

// Open determines if this parser should handle a line and returns the container node if so.
// Returns nil if the line cannot be handled by this parser.
// This method checks if we're within a list context and if the current line matches a list item pattern.
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

// Continue returns how to continue processing this container node.
// For list items, continuation is handled by checking indentation and content.
// This method determines whether the current line should be part of the list item.
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

// Close is called when the container node is being closed.
// For list items, no special cleanup is required.
func (b *fancyListItemParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

// CanInterruptParagraph returns true if this parser can interrupt a paragraph.
// List items can interrupt paragraphs to start new lists.
func (b *fancyListItemParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine returns false as list item parsers don't handle indented lines directly.
// Indented content within list items is handled by other parsers.
func (b *fancyListItemParser) CanAcceptIndentedLine() bool {
	return false
}

// fancyListHTMLRenderer provides HTML rendering for fancy lists.
//
// This renderer generates proper HTML output for lists with custom markers,
// including appropriate CSS classes and start attributes for ordered lists.
// It extends the standard Goldmark HTML renderer to handle fancy list features.
type fancyListHTMLRenderer struct {
	html.Config
}

// RegisterFuncs registers the rendering functions for list nodes.
// This method tells Goldmark to use our custom renderList function for ast.KindList nodes.
func (r *fancyListHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindList, r.renderList)
}

// renderList renders HTML for list elements (both ordered and unordered).
// This method handles the opening and closing of list tags, including special
// handling for fancy list types with custom CSS classes and start attributes.
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
//
// This renderer generates HTML for individual list items within fancy lists.
// Unlike standard list item rendering, this renderer does NOT add value attributes
// to list items, relying instead on the start attribute of the parent list.
type fancyListItemHTMLRenderer struct {
	html.Config
}

// RegisterFuncs registers the rendering functions for list item nodes.
// This method tells Goldmark to use our custom renderListItem function for ast.KindListItem nodes.
func (r *fancyListItemHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindListItem, r.renderListItem)
}

// renderListItem renders HTML for individual list item elements.
// This method creates <li> tags without value attributes, allowing the parent
// list's start attribute and CSS classes to handle proper numbering and styling.
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