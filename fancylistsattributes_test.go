package fancylists

import (
	"fmt"
	"testing"

	"github.com/yuin/goldmark/testutil"
)

var mdattr = CreateGoldmarkInstance(true)

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
}

func TestFancyListsAttributes(t *testing.T) {
	fmt.Println("    Running FancyLists attributes tests...")
	for i, c := range attr_cases {
		testutil.DoTestCase(mdattr, testutil.MarkdownTestCase{
			No:          i,
			Description: c.desc,
			Markdown:    c.md,
			Expected:    c.html,
		}, t)
	}
}
