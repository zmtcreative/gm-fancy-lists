# Goldmark Extension for Pandoc-style Fancy Lists

[![Go Reference](https://pkg.go.dev/badge/github.com/ZMT-Creative/gm-fancy-lists.svg)](https://pkg.go.dev/github.com/ZMT-Creative/gm-fancy-lists)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ZMT-Creative/gm-fancy-lists)
![GitHub License](https://img.shields.io/github/license/ZMT-Creative/gm-fancy-lists)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/ZMT-Creative/gm-fancy-lists)
![GitHub Tag](https://img.shields.io/github/v/tag/ZMT-Creative/gm-fancy-lists?include_prereleases&sort=semver)

This Goldmark extension adds support for Pandoc-style "fancy lists" with extended marker types
including alphabetic markers, roman numeral markers, and hash continuation markers.

## Installation

```bash
go get github.com/ZMT-Creative/gm-fancy-lists
```

## Quick Start

> [!IMPORTANT]
>
> The initialization method has changed starting with v0.2.0. The initialization method below is
> currently the preferred method for enabling the extension.

```go
package main

import (
    "bytes"
    "fmt"

    "github.com/yuin/goldmark"
    fancylists "github.com/ZMT-Creative/gm-fancy-lists"
)

func main() {
    // Create Goldmark instance with fancy lists extension
    md := goldmark.New(
        goldmark.WithExtensions(fancylists.FancyLists),
    )

    // Example markdown with fancy lists
    source := `
1. First numeric item
2. Second numeric item

a. First alphabetic item
b. Second alphabetic item

i. First roman numeral
ii. Second roman numeral
#. Continue previous numbering
#. Another continuation
`

    var buf bytes.Buffer
    if err := md.Convert([]byte(source), &buf); err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
}
```

Alternatively, you can use this initialization method instead:

```go
    // Create Goldmark instance with fancy lists extension
    md := goldmark.New(
        goldmark.WithExtensions(&fancylists.FancyListsOptions{}),
    )
```

> [!NOTE]
>
> These are currently equivalent -- since this extension takes no options (*at this time*), just
> use the first initialization method. In the event that options are later added to the extension,
> the standard `fancylists.FancyLists` initialization method will select the most likely default
> options, while the `&fancylists.FancyListsOptions{}` initialization method will allow you to
> customize these options. But since there currently are no options, this is just redundant for
> now.

## Features

<!-- markdownlint-disable MD033 -->

- **Extended List Markers**: Support for multiple marker types beyond standard numeric lists
  - **Alphabetic Lists**: Lowercase (`a.`, `b.`, `c.`) and uppercase (`A.`, `B.`, `C.`) alphabetic
    markers
  - **Roman Numeral Lists**: Lowercase (`i.`, `ii.`, `iii.`) and uppercase (`I.`, `II.`, `III.`)
    roman numerals <br/>(*see [Roman Numeral Lists](#roman-numeral-lists) below for caveats*)
  - **Hash Continuation**: Use `#.` to continue the current list numbering sequence
- **Automatic Type Detection**: Lists automatically separate when marker types change at the same
  level
- **CSS-Friendly Output**: Generates HTML with specific CSS classes for easy styling
- **Pandoc Compatibility**: Follows Pandoc's fancy list behavior and conventions <br/>(*mostly, see
  [Roman Numeral Lists](#roman-numeral-lists) below for caveats*)

## General Warning

**This extension overrides Goldmark's standard `List` and `ListItem` handling and may potentially
interfere with other extensions that modify list behavior.** It handles both unordered (bullet)
lists and the ordered list types (numbered, alphabetical and roman numerals).

The unit tests for this extension test just this extension alone **and** additional tests are run
with the built-in Goldmark `extension.GFM`, `extension.DefinitionList` and `extension.Footnote`
extensions. We also run tests with the block attributes extension
`github.com/mdigger/goldmark-attributes` (**and the built-in Goldmark extensions**) all enabled and
all tests are currently passing.

This extension is being used in another project that we're working on, using many other commonly
used Goldmark extensions and appears to be working fine.

**However, use with caution in production environments and thoroughly test compatibility with your
specific extension stack.**

## Supported List Types

This section provides a more detailed explanation of what this extension provides and its
limitations. Please read it carefully to make sure you understand the syntax changes and the likely
results.

### Bullet (Unordered List)

No changes have been made to the output of bullet lists. The standard Goldmark handling of
unordered lists has been duplicated in this extension, so unordered lists should appear as they
normally would without this extension. No classes or other attributes are added to the `<ul>`
element.

```markdown
- List item 1
  + list item 1.1
  + list item 1.2
- List item 2
- List item 3
  * list item 3.1
  * list item 3.2
```

### Numeric Lists

Standard numeric lists work as expected:

```markdown
1. First item
2. Second item
3. Third item
```

The common markdown syntax of simply using `1.` for all items at each level also works as it should.

```markdown
1. First item
1. Second item
   1. Subitem 1 of Second item
   1. Subitem 2 of Second item
1. Third item
```

### Alphabetic Lists

Both lowercase and uppercase alphabetic markers:

```markdown
a. First alphabetic item
b. Second alphabetic item

A. First uppercase item
B. Second uppercase item
```

You can also reuse the same character for all entries in the list:

```markdown
1. First number item
1. Second number item

a. First letter item
a. Second letter item
a. Third letter item

i. First roman numeral item
i. Second roman numeral item
i. Third roman numeral item
```

### Hash `#.` Continuation Character

As specified in the Pandoc-style Fancy Lists, after initiating a fancy list with a number, letter
(or the `i/I` letter when starting a roman numeral ordered list), you can use the hash-continuation marker (`#.`) to mark all
subsequent items at that list level. The use of the `#.` continuation character is optional for
number and letter ordered lists.

If you are starting a roman numeral ordered list (using `i/I`, `ii/II`, `iii/III`, or `iv/IV` to
start the list), you must use the `#.` (or `#)`) continuation character for any subsequent list
identifiers that don't begin with `i` (or `I`). Otherwise you will get unexpected results.

### Roman Numeral Lists

Roman numeral lists are a special case that deviates slightly from Pandoc. We **ONLY** accept the roman
numeral numbers `1-4` (i.e., `i`, `ii`, `iii` and `iv` and the uppercase equivalents) to indicate
the **start** of a roman numeral ordered list. This simplifies the parsing of lists, so we don't
have to figure out if a list starting with a `c` is an alphabetic list (start value = `3`) or a
roman numeral list (start value = `100`).

It is assumed that most people don't need to start a roman numeral ordered list with some arbitrary
large roman numeral (like `MMXXV` for `2025`). This was a design decision on our part to keep roman
numeral handling relatively simple.

These are valid ordered lists using roman numerals:

```markdown
i. First roman numeral item
ii. Second roman numeral item
iii. Third roman numeral item

IV. First uppercase roman (starts with roman numeral 4)
#. Second uppercase roman (must use `#.` continuation here)

  i. item one
 ii. item two
iii. item three
 iv. item four
```

The following is **NOT** a valid roman numeral list (*the `v.` will NOT be recognized as a roman numeral here*):

```markdown
  i. item one
 ii. item two
iii. item three
 iv. item four
  v. item five
 vi. item six
vii. item seven
```

...will produce this output:

```text
  i. item one
 ii. item two
iii. item three
 iv. item four


  v. item five
  w. item six
  x. item seven
```

This happens because the identifier `v.` is not recognized as a roman numeral and is instead
interpreted as an alphabetic character. The extension therefore assumes this is a new ordered list
starting with the alphabetic character `v` and increments the new list accordingly.

## HTML Output

The extension generates HTML with CSS classes for easy styling of **ordered** lists:

- **Numeric lists**: `class="fancy fl-num"`
- **Lowercase alphabetic**: `class="fancy fl-lcalpha"`
- **Uppercase alphabetic**: `class="fancy fl-ucalpha"`
- **Lowercase roman**: `class="fancy fl-lcroman"`
- **Uppercase roman**: `class="fancy fl-ucroman"`

All ordered lists include appropriate `type` and `start` attributes. List items do not
include `value` attributes, allowing the browser to handle numbering naturally.

## CSS Styling Example

```css
/* Style different list types */
ol.fl-lcalpha { list-style-type: lower-alpha; }
ol.fl-ucalpha { list-style-type: upper-alpha; }
ol.fl-lcroman { list-style-type: lower-roman; }
ol.fl-ucroman { list-style-type: upper-roman; }
ol.fl-num { list-style-type: decimal; }

/* Add custom styling for fancy lists */
ol.fancy {
    margin: 1em 0;
    padding-left: 2em;
}
```

> [!NOTE]
>
> Styling using these classes is optional. The default browsers styling should be adequate for most usage.

## List Type Changes

When list marker types change at the same level, the current list automatically closes and a new list begins:

```markdown
1. Numeric item
2. Another numeric item
a. This starts a new alphabetic list
b. Continues the alphabetic list
I. This starts a new roman list
```

This generates three separate `<ol>` elements with different `type` attributes and CSS classes:

```html
<ol class="fancy fl-num" type="1" start="1">
   <li>Numeric item</li>
   <li>Another numeric item</li>
</ol>
<ol class="fancy fl-lcalpha" type="a" start="1">
   <li>This starts a new alphabetic list</li>
   <li>Continues the alphabetic list</li>
</ol>
<ol class="fancy fl-ucroman" type="I" start="1">
   <li>This starts a new uppercase roman list</li>
</ol>
```

> [!TIP]
>
> **Don't use this feature if you can avoid it!**
>
> **It is better to separate new ordered lists with a blank line.** Depending on this list-type
> change feature parsing correctly (*instead of following general CommonMark list handling syntax*)
> can lead to strange quirks in the way your Markdown is parsed, especially with roman numerals.
>
> This feature is here to (*mostly*) match the Pandoc-style handling, but it is **not** a perfect
> match and thus should be avoided if possible in your Markdown files.

### Special Considerations for Type Changes with Roman Numerals

You **cannot** make a list-type change if your current list type is lowercase alphabetic and you
use a lowercase `i` to indicate changing to a new lowercase roman numeral ordered list. The parser
will ignore this and treat the `i` as just another letter identifier in the current list.

The same is true in reverse -- if you are in an uppercase alphabetic list and you use an uppercase
`I` expecting to start a new uppercase roman numeral list, the uppercase `I` will be treated as
just another uppercase letter identifier in the current list.

If you use an uppercase `I` in a current lowercase lettered ordered list (or a lowercase `i` in an
uppercase lettered ordered list), the parser **WILL** detect the case change and make the
transition to the new roman numeral list.

If you place an uppercase or lowercase `i/I` roman numeral in an existing number ordered list, the
parser will correctly transition to the new roman numeral list.

## Dependencies

- [Goldmark](https://github.com/yuin/goldmark) - The CommonMark-compliant Markdown parser
- [Roman Numeral](https://github.com/brandenc40/romannumeral) - Roman numeral conversion utilities

## Compatibility Notes

- **Goldmark Version**: Tested with Goldmark v1.7.13
- **Go Version**: Requires Go 1.22 or later
- **Extension Conflicts**: May conflict with other extensions that override list parsing behavior
- **Standard Compliance**: Extends CommonMark specification following Pandoc conventions

## License

See [LICENSE](LICENSE.md) file for details.
