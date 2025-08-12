package fancylists

import (
	"testing"

	"github.com/fatih/color"
	blockattr "github.com/mdigger/goldmark-attributes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/testutil"
)

// var markdown = goldmark.New(
// 	goldmark.WithExtensions(
// 		&FancyLists{},
// 	),
// )

var markdown = CreateGoldmarkInstance(createOptions{
	blockAttributes: false,
	enableGFM:       false,
})

var mdGFM = CreateGoldmarkInstance(createOptions{
	blockAttributes: false,
	enableGFM:       true,
})

type TestCase struct {
	desc string
	md   string
	html string
}

var cases = [...]TestCase{
	{
		desc: "Invalid Ordered and Unordered lists (missing space between marker and content)",
		md:   `-one

2.two`,
		html: `<p>-one</p>
<p>2.two</p>`},
	{
		desc: "Simple Unordered List with '-'",
		md:   `- First item
- Second item
- Third item
`,
		html: `<ul>
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>`},
	{
		desc: "Unordered list starting with one blank line",
		md:   `-
  foo`,
		html: `<ul>
<li>foo</li>
</ul>`},
	{
		desc: "Unordered list starting with more than one blank line",
		md:   `-

  foo`,
		html: `<ul>
<li></li>
</ul>
<p>foo</p>`},
	{
		desc: "Unordered list starting with one blank line, and\n  both indented and fenced code blocks",
		md:   `-
  foo
-
  ` + "```" + `
  bar
  ` + "```" + `
-
      baz`,
		html: `<ul>
<li>foo</li>
<li>
<pre><code>bar
</code></pre>
</li>
<li>
<pre><code>baz
</code></pre>
</li>
</ul>`},
	{
		desc: "Simple Unordered List with '+'",
		md:   `+ First item
+ Second item
+ Third item
`,
		html: `<ul>
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>`},
	{
		desc: "Simple Unordered List with '*'",
		md:   `* First item
* Second item
* Third item
`,
		html: `<ul>
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ul>`},
	{
		desc: "Simple Unordered List with an empty item",
		md:   `- foo
-
- bar`,
		html: `<ul>
<li>foo</li>
<li></li>
<li>bar</li>
</ul>`},
	{
		desc: "Unordered List with incorrect indentation of continuation text",
		md:   `- one

 two`,
		html: `<ul>
<li>one</li>
</ul>
<p>two</p>`},
	{
		desc: "Unordered List with code-block indent",
		md:   ` -    one

     two`,
		html: `<ul>
<li>one</li>
</ul>
<pre><code> two
</code></pre>`},
	{
		desc: "Simple Ordered List with numbers",
		md:   `1. First item
2. Second item
3. Third item
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with empty second item",
		md:   `1. foo
2.
3. bar`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>foo</li>
<li></li>
<li>bar</li>
</ol>`},
	{
		desc: "Simple Ordered List with same number",
		md:   `1. First item
1. Second item
1. Third item
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with same letter (lowercase)",
		md:   `a. First item
a. Second item
a. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with same roman numeral (lowercase)",
		md:   `i. First item
i. Second item
i. Third item
`,
		html: `<ol class="fancy fl-lcroman" type="i" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with lower roman numeral in second and third item (lowercase)",
		md:   `ii. First item
i. Second item
i. Third item
`,
		html: `<ol class="fancy fl-lcroman" type="i" start="2">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with number and hash",
		md:   `1. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with letters (lowercase)",
		md:   `a. First item
b. Second item
c. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with same letter (lowercase)",
		md:   `a. First item
a. Second item
a. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with letter and hash (lowercase)",
		md:   `a. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with letter and hash (uppercase)",
		md:   `A. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-ucalpha" type="A" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with first 4 roman numerals (lowercase)",
		md:   `  i. First item
 ii. Second item
iii. Third item
 iv. Fourth item
`,
		html: `<ol class="fancy fl-lcroman" type="i" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
<li>Fourth item</li>
</ol>`},
	{
		desc: "Simple Ordered List with first seven roman numeral (lowercase)",
		md:   `  i. First item
 ii. Second item
iii. Third item
 iv. Fourth item
  v. Fifth item
 vi. Sixth item
vii. Seventh item
`,
		html: `<ol class="fancy fl-lcroman" type="i" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
<li>Fourth item</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="22">
<li>Fifth item</li>
<li>Sixth item</li>
<li>Seventh item</li>
</ol>`},
	{
		desc: "Ordered List with roman numeral NOT beginning with 'i' (treated as alphabetic)",
		md:   `vi. First item
vii. Second item
#. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a" start="581">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with roman numeral (uppercase)",
		md:   `I. First item
II. Second item
III. Third item
`,
		html: `<ol class="fancy fl-ucroman" type="I" start="1">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Simple Ordered List with roman numeral (uppercase) starting at IV",
		md:   `IV. First item
#. Second item
#. Third item
`,
		html: `<ol class="fancy fl-ucroman" type="I" start="4">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Ordered List with numbers starting at 8",
		md:   `8. First item
9. Second item
10. Third item
`,
		html: `<ol class="fancy fl-num" type="1" start="8">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Ordered List with letters starting at g (lowercase)",
		md:   `g. First item
h. Second item
i. Third item
`,
		html: `<ol class="fancy fl-lcalpha" type="a" start="7">
<li>First item</li>
<li>Second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Ordered List two levels",
		md:   `1. First item
#. Second item
   A. Subitem 2.1
   A. Subitem 2.2
   #. Subitem 2.3
#. Third item
   ii. Subitem 3.1
   #. Subitem 3.2
#. Fourth item
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>First item</li>
<li>Second item
<ol class="fancy fl-ucalpha" type="A" start="1">
<li>Subitem 2.1</li>
<li>Subitem 2.2</li>
<li>Subitem 2.3</li>
</ol>
</li>
<li>Third item
<ol class="fancy fl-lcroman" type="i" start="2">
<li>Subitem 3.1</li>
<li>Subitem 3.2</li>
</ol>
</li>
<li>Fourth item</li>
</ol>`},
	{
		desc: "Simple Ordered List with numbers and multi-line item 2",
		md:   `1. First item
2. Second item

   Continuation of second item
3. Third item
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>
<p>First item</p>
</li>
<li>
<p>Second item</p>
<p>Continuation of second item</p>
</li>
<li>
<p>Third item</p>
</li>
</ol>`},
	{
		desc: "Simple Ordered List with numbers and compact multi-line item 2",
		md:   `1. First item
2. Second item
   Continuation of second item
3. Third item
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>First item</li>
<li>Second item
Continuation of second item</li>
<li>Third item</li>
</ol>`},
	{
		desc: "Ordered list with block elements (indented code and blockquote)",
		md: `1.  A paragraph
    with two lines.

        indented code

    > A block quote.`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>
<p>A paragraph
with two lines.</p>
<pre><code>indented code
</code></pre>
<blockquote>
<p>A block quote.</p>
</blockquote>
</li>
</ol>`},
	{
		desc: "Ordered list inside blockquotes",
		md: `   > > 1.  one
>>
>>     two`,
		html: `<blockquote>
<blockquote>
<ol class="fancy fl-num" type="1" start="1">
<li>
<p>one</p>
<p>two</p>
</li>
</ol>
</blockquote>
</blockquote>`},
	{
		desc: "Unordered list inside blockquotes",
		md: `>>- one
>>
  >  > two`,
		html: `<blockquote>
<blockquote>
<ul>
<li>one</li>
</ul>
<p>two</p>
</blockquote>
</blockquote>`},
	{
		desc: "Indented code block inside unordered list",
		md: `- Foo

      bar


      baz`,
		html: `<ul>
<li>
<p>Foo</p>
<pre><code>bar


baz
</code></pre>
</li>
</ul>`},
	{
		desc: "Ordered List: Valid number marker",
		md: `123456789. ok`,
		html: `<ol class="fancy fl-num" type="1" start="123456789">
<li>ok</li>
</ol>`},
	{
		desc: "Ordered List: Invalid number marker",
		md: `1234567890. not ok`,
		html: `<p>1234567890. not ok</p>`},
	{
		desc: "Ordered List: Marker using 0",
		md: `0. ok`,
		html: `<ol class="fancy fl-num" type="1" start="0">
<li>ok</li>
</ol>`},
	{
		desc: "Ordered List: Marker using 003",
		md: `003. ok`,
		html: `<ol class="fancy fl-num" type="1" start="3">
<li>ok</li>
</ol>`},
	{
		desc: "Ordered List: Invalid negative number marker",
		md: `-1. not ok`,
		html: `<p>-1. not ok</p>`},
	{
		desc: "Empty Lists cannot interrupt a paragraph",
		md: `foo
*

foo
1.`,
		html: `<p>foo
*</p>
<p>foo
1.</p>`},
	{
		desc: "Unordered List - sublists need two space indents",
		md: `- foo
  - bar
    - baz
      - boo`,
		html: `<ul>
<li>foo
<ul>
<li>bar
<ul>
<li>baz
<ul>
<li>boo</li>
</ul>
</li>
</ul>
</li>
</ul>
</li>
</ul>`},
	{
		desc: "Unordered List - single space indents are NOT sublists",
		md: `- foo
 - bar
  - baz
   - boo`,
		html: `<ul>
<li>foo</li>
<li>bar</li>
<li>baz</li>
<li>boo</li>
</ul>`},
	{
		desc: "Unordered List inside Ordered List \n  - indents must account for parent list item indent",
		md: `10) foo
    - bar`,
		html: `<ol class="fancy fl-num" type="1" start="10">
<li>foo
<ul>
<li>bar</li>
</ul>
</li>
</ol>`},
	{
		desc: "Unordered List inside Ordered List \n  - indents must account for parent list item indent \n  - three is not enough here",
		md: `10) foo
   - bar`,
		html: `<ol class="fancy fl-num" type="1" start="10">
<li>foo</li>
</ol>
<ul>
<li>bar</li>
</ul>`},
	{
		desc: "A list item can contain a heading",
		md: `- # Foo
- Bar
  ---
  baz`,
		html: `<ul>
<li>
<h1>Foo</h1>
</li>
<li>
<h2>Bar</h2>
baz</li>
</ul>`},
	{
		desc: "A Basic Fancylist OrderedList Test",
		md: `1. foo 1
#. foo 2
A. bar A
#. bar B`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>foo 1</li>
<li>foo 2</li>
</ol>
<ol class="fancy fl-ucalpha" type="A" start="1">
<li>bar A</li>
<li>bar B</li>
</ol>`},
	{
		desc: "A Multilevel Fancylist OrderedList Test",
		md: `1. foo 1
#. foo 2
   a. baz 'a'
   b. baz 'b'
   A. boo 'A'
   B. boo 'B'
#. foo 3
A. bar A
A. bar B
   iii. boo 'iii'
   #.   boo 'iv'
   #.   boo 'v'
   #.   boo 'vi'
   I.   booboo 'I'
   #.   booboo 'II'
A. bar C`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>foo 1</li>
<li>foo 2
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>baz 'a'</li>
<li>baz 'b'</li>
</ol>
<ol class="fancy fl-ucalpha" type="A" start="1">
<li>boo 'A'</li>
<li>boo 'B'</li>
</ol>
</li>
<li>foo 3</li>
</ol>
<ol class="fancy fl-ucalpha" type="A" start="1">
<li>bar A</li>
<li>bar B
<ol class="fancy fl-lcroman" type="i" start="3">
<li>boo 'iii'</li>
<li>boo 'iv'</li>
<li>boo 'v'</li>
<li>boo 'vi'</li>
</ol>
<ol class="fancy fl-ucroman" type="I" start="1">
<li>booboo 'I'</li>
<li>booboo 'II'</li>
</ol>
</li>
<li>bar C</li>
</ol>`},
	{
		desc: "A full Fancylist Mixed List Test",
		md: `1. foo 1
2. foo 2
   i. bar roman 'i'
   #. bar roman 'ii'
   #. bar roman 'iii'
      - bullet item 1
      - bullet item 2
   #. bar roman 'vi'
   #. bar roman 'v'
#. foo 3
#. foo 4
   j. boo alpha 'j'
   #. boo alpha 'k'
      a. boobaz alpha k.a
      b. boobaz alpha k.b
      z. boobaz alpha k.c
   #. boo alpha 'l'
C. foofoo C
#. foofoo D
   1) foofoo sub B.1
   #) foofoo sub B.2
   5) foofoo sub B.3
#. foofoo E`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>foo 1</li>
<li>foo 2
<ol class="fancy fl-lcroman" type="i" start="1">
<li>bar roman 'i'</li>
<li>bar roman 'ii'</li>
<li>bar roman 'iii'
<ul>
<li>bullet item 1</li>
<li>bullet item 2</li>
</ul>
</li>
<li>bar roman 'vi'</li>
<li>bar roman 'v'</li>
</ol>
</li>
<li>foo 3</li>
<li>foo 4
<ol class="fancy fl-lcalpha" type="a" start="10">
<li>boo alpha 'j'</li>
<li>boo alpha 'k'
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>boobaz alpha k.a</li>
<li>boobaz alpha k.b</li>
<li>boobaz alpha k.c</li>
</ol>
</li>
<li>boo alpha 'l'</li>
</ol>
</li>
</ol>
<ol class="fancy fl-ucalpha" type="A" start="3">
<li>foofoo C</li>
<li>foofoo D
<ol class="fancy fl-num" type="1" start="1">
<li>foofoo sub B.1</li>
<li>foofoo sub B.2</li>
<li>foofoo sub B.3</li>
</ol>
</li>
<li>foofoo E</li>
</ol>`},
	{
		desc: "A full Fancylist Test -- Roman Numerals that don't start with 'i' are treated as alphabetic instead of roman numerals",
		md: `1. foo 1
2. foo 2
   vi. bar roman 'vi'
   #. bar roman 'vj'
   #. bar roman 'vk'
      - bullet item 1
      - bullet item 2
   #. bar roman 'vl'
   #. bar roman 'vm'
#. foo 3`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>foo 1</li>
<li>foo 2
<ol class="fancy fl-lcalpha" type="a" start="581">
<li>bar roman 'vi'</li>
<li>bar roman 'vj'</li>
<li>bar roman 'vk'
<ul>
<li>bullet item 1</li>
<li>bullet item 2</li>
</ul>
</li>
<li>bar roman 'vl'</li>
<li>bar roman 'vm'</li>
</ol>
</li>
<li>foo 3</li>
</ol>`},
	{
		desc: "A paragraph between lists creates two separate lists and hashes are consider numeric here",
		md: `1. First item
2. Second item

Some text here.

#. Third item (continues from 3)
#. Fourth item (continues from 4)`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>First item</li>
<li>Second item</li>
</ol>
<p>Some text here.</p>
<ol class="fancy fl-num" type="1" start="1">
<li>Third item (continues from 3)</li>
<li>Fourth item (continues from 4)</li>
</ol>`},
	{
		desc: "A mixed list with different types that should create three separate ordered lists\n (number, lcalpha and ucalpha)",
		md: `1. Numeric item
2. Another numeric item
a. This starts a new alphabetic list
b. Continues the alphabetic list
A. This starts a new uppercase alpha list
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>Numeric item</li>
<li>Another numeric item</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>This starts a new alphabetic list</li>
<li>Continues the alphabetic list</li>
</ol>
<ol class="fancy fl-ucalpha" type="A" start="1">
<li>This starts a new uppercase alpha list</li>
</ol>`},
	{
		desc: "A mixed list with different types that should create three separate ordered lists \n (number, lcalpha and lcroman)",
		md: `1. Numeric item
2. Another numeric item
a. This starts a new alphabetic list
b. Continues the alphabetic list
i. This continues the lowercase alphabetic list
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>Numeric item</li>
<li>Another numeric item</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>This starts a new alphabetic list</li>
<li>Continues the alphabetic list</li>
<li>This continues the lowercase alphabetic list</li>
</ol>`},
	{
		desc: "A mixed list with different types that should create three separate ordered lists \n (number, lcroman and lcalpha)",
		md: `1. Numeric item
2. Another numeric item
i. This starts a new lowercase roman list
a. This starts a new alphabetic list
b. Continues the alphabetic list
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>Numeric item</li>
<li>Another numeric item</li>
</ol>
<ol class="fancy fl-lcroman" type="i" start="1">
<li>This starts a new lowercase roman list</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>This starts a new alphabetic list</li>
<li>Continues the alphabetic list</li>
</ol>
`},
	{
		desc: "A mixed list with different types that should create three separate ordered lists \n (number, lcalpha and ucroman)",
		md: `1. Numeric item
2. Another numeric item
a. This starts a new alphabetic list
b. Continues the alphabetic list
I. This starts a new uppercase roman list
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>Numeric item</li>
<li>Another numeric item</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>This starts a new alphabetic list</li>
<li>Continues the alphabetic list</li>
</ol>
<ol class="fancy fl-ucroman" type="I" start="1">
<li>This starts a new uppercase roman list</li>
</ol>`},
	{
		desc: "A mixed list with different types that should create three separate ordered lists \n (number, ucalpha and lcroman)",
		md: `1. Numeric item
2. Another numeric item
A. This starts a new alphabetic list
B. Continues the alphabetic list
i. This starts a new lowercase roman list
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>Numeric item</li>
<li>Another numeric item</li>
</ol>
<ol class="fancy fl-ucalpha" type="A" start="1">
<li>This starts a new alphabetic list</li>
<li>Continues the alphabetic list</li>
</ol>
<ol class="fancy fl-lcroman" type="i" start="1">
<li>This starts a new lowercase roman list</li>
</ol>`},
	{
		desc: "A mixed list with different types that should create three separate ordered lists \n (number, ucroman and lcalpha)",
		md: `1. Numeric item
2. Another numeric item
I. This starts a new uppercase roman list
a. This starts a new alphabetic list
b. Continues the alphabetic list
`,
		html: `<ol class="fancy fl-num" type="1" start="1">
<li>Numeric item</li>
<li>Another numeric item</li>
</ol>
<ol class="fancy fl-ucroman" type="I" start="1">
<li>This starts a new uppercase roman list</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="1">
<li>This starts a new alphabetic list</li>
<li>Continues the alphabetic list</li>
</ol>
`},
}

func TestFancyLists(t *testing.T) {
	color.Cyan("  + Running Basic FancyLists tests\n      (all Goldmark Extensions disabled)...\n")
	for i, c := range cases {
		testutil.DoTestCase(markdown, testutil.MarkdownTestCase{
			No:          i,
			Description: c.desc,
			Markdown:    c.md,
			Expected:    c.html,
		}, t)
	}
}

func TestFancyListsGFM(t *testing.T) {
	color.Green("  + Running Basic (same) FancyLists tests \n      with GFM and PHP Markdown Extensions enabled...\n")
	for i, c := range cases {
		testutil.DoTestCase(markdown, testutil.MarkdownTestCase{
			No:          i,
			Description: c.desc,
			Markdown:    c.md,
			Expected:    c.html,
		}, t)
	}
}

type createOptions struct {
	blockAttributes bool
	enableGFM       bool
}

// CreateGoldmarkInstance creates and configures a new Goldmark instance.
// The options parameter allows for customization of the instance.
func CreateGoldmarkInstance(opt createOptions) goldmark.Markdown {
	// Initialize a new Goldmark instance with default options for testing fancylists.
    options := []goldmark.Option{
        goldmark.WithParserOptions(),
        goldmark.WithExtensions(
			&FancyLists{},
        ),
    }

	// Enable GitHub Flavored Markdown (GFM) extensions if requested.
	// Also enables PHP Markdown Definition List and Footnote extensions.
	if opt.enableGFM {
		options = append(options,
			goldmark.WithExtensions(
				extension.GFM,
				extension.DefinitionList,
				extension.Footnote,
				// extension.Typographer,
			),
		)
	}

	// Enable github.com/mdigger/goldmark-attributes if requested.
	// currently used by fancylistsattributes_test.go
	if opt.blockAttributes {
		options = append(options,
			blockattr.Enable,
			goldmark.WithParserOptions(
	            parser.WithAutoHeadingID(), // Automatically generate IDs for headings
				parser.WithAttribute(),
			),
		)
	}

    return goldmark.New(options...)
}
