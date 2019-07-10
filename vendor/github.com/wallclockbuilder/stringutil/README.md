#stringutil [![Build Status](https://travis-ci.org/wallclockbuilder/stringutil.svg)](https://travis-ci.org/wallclockbuilder/stringutil) [![GoDoc](https://godoc.org/github.com/wallclockbuilder/stringutil?status.svg)](http://godoc.org/github.com/wallclockbuilder/stringutil) [![Readability](https://img.shields.io/badge/cyclomatic%20complexity-5-brightgreen.svg)](http://locometer.com/github.com/wallclockbuilder/stringutil)
Ruby's string manipulation magic brought to Golang

Why?
>  "What was the biggest surprise you encountered rolling out Go?" I knew the answer instantly: Although we expected C++ programmers to see Go as an alternative, instead **most Go programmers come from languages like Python and Ruby**.
 Robert Pike, [Less is exponentially more](http://commandcenter.blogspot.com/2012/06/less-is-exponentially-more.html)

stringutil makes all the string manipulation methods from ruby accessible in go.


##Install
```bash
 go get https://github.com/wallclockbuilder/stringutil
```


##Import
```golang
 import "stringutil" "https://github.com/wallclockbuilder/stringutil"
```


## Example
I want to remove the trailing space
 from the end of a string.

This is what Go forces me to do:
```golang
strings.TrimRightFunc(s, unicode.IsSpace)
```
This is the simple version from Ruby:
```golang
stringutil.Rstrip(s)
```
I prefer the Ruby version.


##Use
```golang
package main
import "github.com/wallclockbuilder/stringutil"

func main() {
  stringutil.Capitalize("abcde")          #=> "ABCDE"
  stringutil.Reverse("stressed")          #=> "desserts"
  stringutil.Swapcase("Hello")            #=> "hELLO"
}
```


##Documentation
`go doc` style documentation for this package is available online at http://godoc.org/github.com/wallclockbuilder/stringutil


## Contributing
1. Fork it ( https://github.com/wallclockbuilder/stringutil )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
