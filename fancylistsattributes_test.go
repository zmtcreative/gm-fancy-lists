package fancylists

import (
	"testing"

	"github.com/fatih/color"
	"github.com/yuin/goldmark/testutil"
)

var mdattr = CreateGoldmarkInstance(createOptions{
	blockAttributes: true,
	enableGFM:       true,
})

type TestCaseAttributes struct {
	desc string
	md   string
	html string
}

var attr_cases = [...]TestCaseAttributes{
	{
		desc: "ATTR: Invalid Ordered and Unordered lists (missing space between marker and content)",
		md:   `-one

2.two`,
		html: `<p>-one</p>
<p>2.two</p>`},
	{
		desc: "ATTR: Simple Unordered List with '-' and {.sbs} class attribute",
		md:   `- First item
- Second item
- Third item
{.sbs}
`,
		html: `<ul class="sbs">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>`},
	{
		desc: "ATTR: Simple Ordered List with numbers and {.sbs} class attribute",
		md:   `1. First item
2. Second item
3. Third item
{.sbs}
`,
		html: `<ol class="fancy fl-num sbs" type="1" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "ATTR: Simple Unordered List with '-' and {.foo} class attribute",
		md:   `- First item
- Second item
- Third item
{.foo}
`,
		html: `<ul class="foo">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>`},
	{
		desc: "ATTR: Simple Ordered List with numbers and {.foo} class attribute",
		md:   `1. First item
2. Second item
3. Third item
{.foo}
`,
		html: `<ol class="fancy fl-num foo" type="1" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: `ATTR: Simple Unordered List with '-' and {.foo} class attribute with bar="baz" custom attribute`,
		md:   `- First item
- Second item
- Third item
{.foo bar="baz"}
`,
		html: `<ul class="foo" bar="baz">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>`},
	{
		desc: `ATTR: Simple Ordered List with numbers and {.foo} class attribute with bar="baz" custom attribute`,
		md:   `1. First item
2. Second item
3. Third item
{.foo bar="baz"}
`,
		html: `<ol class="fancy fl-num foo" type="1" start="1" bar="baz">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "ATTR: Multi-Level Unordered List with {.foo} class attribute on level 1 and {.bar} class attribute on level 2",
		md:   `- First item
- Second item
  + Subitem one
  + Subitem two
    * Subsubitem one
	* Subsubitem two
  + Subitem three
  + Subitem four
  {.baz}
- Third item
{.foo}
`,
		html: `<ul class="foo">
<li>First item</li>
<li>Second item
<ul class="baz">
<li>Subitem one</li>
<li>Subitem two
<ul>
<li>Subsubitem one</li>
<li>Subsubitem two</li>
</ul>
</li>
<li>Subitem three</li>
<li>Subitem four</li>
</ul>
</li>
<li>Third item</li>
</ul>`},
	{
		desc: "ATTR: Multi-Level Ordered List with {.foo} class attribute on level 1 and {.baz} class attribute on level 2",
		md:   `1. First item
2. Second item
   1. Subitem one
   2. Subitem two
   {.baz}
3. Third item
{.foo}
`,
		html: `<ol class="fancy fl-num foo" type="1" start="1">
<li>First item</li>
<li>Second item
<ol class="fancy fl-num baz" type="1" start="1">
<li>Subitem one</li>
<li>Subitem two</li>
</ol>
</li>
<li>Third item</li>
</ol>
`},
}

func TestFancyListsAttributes(t *testing.T) {
	color.HiCyan("  + Running more FancyLists tests with goldmark-attributes enabled and \n      with GFM and PHP Markdown Extensions enabled...\n")
	for i, c := range attr_cases {
		testutil.DoTestCase(mdattr, testutil.MarkdownTestCase{
			No:          i,
			Description: c.desc,
			Markdown:    c.md,
			Expected:    c.html,
		}, t)
	}
}
