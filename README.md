# golang-writing-web-applications

Fork of official <https://golang.org/doc/articles/wiki/> example code.

- [golang-writing-web-applications](#golang-writing-web-applications)
  - [Summary](#summary)
  - [Future Development](#future-development)
  - [References](#references)

## Summary

This is a fork I created of of the official
<https://golang.org/doc/articles/wiki/> example code while following along
with the article. Many thanks to the authors and contributors to that article
as I found it a very useful learning exercise.

Further development has been done towards the *Other tasks* items included at
the end of that article:

> - Store templates in `tmpl/` and page data in `data/`.
> - Add a handler to make the web root redirect to `/view/FrontPage`.
> - Spruce up the page templates by making them valid HTML and adding some CSS
>   rules.
> - Implement inter-page linking by converting instances of `[PageName]` to
>   `<a href="/view/PageName">PageName</a>`. (hint: you could use
>   `regexp.ReplaceAllFunc` to do this)

See the issues list, commits or current state of the code for my solution to
(some of) these items. I'm still fairly new to Go, so feel free to open an
issue or give me a shout via Twitter for anything I'm doing wrong.

## Future Development

Aside from the post-article learning exercises listed above and some minor
tweaks of my own (e.g., use a popular Markdown to HTML parser package), this
code is as-is and I do not plan on making further changes it.

I'm setting this repo's visibility to public in the hope that others could
benefit from what I've learned (e.g., Google search), but moving on to other
projects in an attempt to further my Go knowledge.

## References

- <https://golang.org/doc/articles/wiki/>

- <https://github.com/microcosm-cc/bluemonday>
- <https://github.com/russross/blackfriday>
