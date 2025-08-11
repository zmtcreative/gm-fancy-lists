package fancylists

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(
		&FancyLists{},
	),
)

type TestCase struct {
	desc string
	md   string
	html string
}

var cases = [...]TestCase{
	{
		desc: "Simple Ordered List with numbers",
		md:   `1. First item
2. Second item
3. Third item
`,
		html: `<ol class="fancy fl-num" type="1">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with same number",
		md:   `1. First item
1. Second item
1. Third item
`,
		html: `<ol class="fancy fl-num" type="1">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with number and hash",
		md:   `1. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-num" type="1">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with letters (lowercase)",
		md:   `a. First item
b. Second item
c. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with same letter (lowercase)",
		md:   `a. First item
a. Second item
a. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with letter and hash (lowercase)",
		md:   `a. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with letter and hash (uppercase)",
		md:   `A. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-ucalpha" type="A">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with roman numeral (lowercase)",
		md:   `i. First item
ii. Second item
iii. Third item
`,
		html: `<ol class="fancy fl-lcroman" type="i">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with roman numeral (uppercase)",
		md:   `I. First item
II. Second item
III. Third item
`,
		html: `<ol class="fancy fl-ucroman" type="I">
<li value="1">First item</li>
<li value="2">Second item</li>
<li value="3">Third item</li>
</ol>`},
	{
		desc: "Ordered List with numbers starting at 8",
		md:   `8. First item
9. Second item
10. Third item
`,
		html: `<ol class="fancy fl-num" type="1">
<li value="8">First item</li>
<li value="9">Second item</li>
<li value="10">Third item</li>
</ol>`},
	{
		desc: "Ordered List with letters starting at g (lowercase)",
		md:   `g. First item
h. Second item
i. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a">
<li value="7">First item</li>
<li value="8">Second item</li>
<li value="9">Third item</li>
</ol>`},
	{
		desc: "Ordered List two levels",
		md:   `1. First item
#. Second item
   A. Subitem 2.1
   A. Subitem 2.2
   #. Subitem 2.3
#. Third item
   i. Subitem 3.1
   i. Subitem 3.2
#. Fourth item
`,
		html: `<ol class="fancy fl-num" type="1">
<li value="1">First item</li>
<li value="2">Second item
<ol class="fancy fl-ucalpha" type="A">
<li value="1">Subitem 2.1</li>
<li value="2">Subitem 2.2</li>
<li value="3">Subitem 2.3</li>
</ol>
</li>
<li value="3">Third item
<ol class="fancy fl-lcroman" type="i">
<li value="1">Subitem 3.1</li>
<li value="2">Subitem 3.2</li>
</ol>
</li>
<li value="4">Fourth item</li>
</ol>`},
}

func TestFancyLists(t *testing.T) {
	for i, c := range cases {
		testutil.DoTestCase(markdown, testutil.MarkdownTestCase{
			No:          i,
			Description: c.desc,
			Markdown:    c.md,
			Expected:    c.html,
		}, t)
	}
}