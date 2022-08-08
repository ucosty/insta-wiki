// Code generated by qtc from "index.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line templates/index.qtpl:1
package templates

//line templates/index.qtpl:1
import "github.com/ucosty/insta-wiki/models"

//line templates/index.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line templates/index.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line templates/index.qtpl:3
func StreamIndexContent(qw422016 *qt422016.Writer, page *models.Page) {
//line templates/index.qtpl:3
	qw422016.N().S(`
<div class="content wrap">
    <nav class="tabs">
        <a href="/wiki/`)
//line templates/index.qtpl:6
	qw422016.E().S(page.Path)
//line templates/index.qtpl:6
	qw422016.N().S(`/history">History</a>
        <a href="/wiki/`)
//line templates/index.qtpl:7
	qw422016.E().S(page.Path)
//line templates/index.qtpl:7
	qw422016.N().S(`/edit">Edit</a>
        <a href="/wiki/`)
//line templates/index.qtpl:8
	qw422016.E().S(page.Path)
//line templates/index.qtpl:8
	qw422016.N().S(`" class="selected">Read</a>
    </nav>

    <h1>`)
//line templates/index.qtpl:11
	qw422016.E().S(page.Title)
//line templates/index.qtpl:11
	qw422016.N().S(`</h1>
    <p>`)
//line templates/index.qtpl:12
	qw422016.N().S(page.BodyFormatted)
//line templates/index.qtpl:12
	qw422016.N().S(`</p>
</div>

<script>hljs.highlightAll();</script>
`)
//line templates/index.qtpl:16
}

//line templates/index.qtpl:16
func WriteIndexContent(qq422016 qtio422016.Writer, page *models.Page) {
//line templates/index.qtpl:16
	qw422016 := qt422016.AcquireWriter(qq422016)
//line templates/index.qtpl:16
	StreamIndexContent(qw422016, page)
//line templates/index.qtpl:16
	qt422016.ReleaseWriter(qw422016)
//line templates/index.qtpl:16
}

//line templates/index.qtpl:16
func IndexContent(page *models.Page) string {
//line templates/index.qtpl:16
	qb422016 := qt422016.AcquireByteBuffer()
//line templates/index.qtpl:16
	WriteIndexContent(qb422016, page)
//line templates/index.qtpl:16
	qs422016 := string(qb422016.B)
//line templates/index.qtpl:16
	qt422016.ReleaseByteBuffer(qb422016)
//line templates/index.qtpl:16
	return qs422016
//line templates/index.qtpl:16
}

//line templates/index.qtpl:18
func StreamIndex(qw422016 *qt422016.Writer, page *models.Page) {
//line templates/index.qtpl:18
	qw422016.N().S(`
`)
//line templates/index.qtpl:19
	StreamBase(qw422016, page.Title, IndexContent(page))
//line templates/index.qtpl:19
	qw422016.N().S(`
`)
//line templates/index.qtpl:20
}

//line templates/index.qtpl:20
func WriteIndex(qq422016 qtio422016.Writer, page *models.Page) {
//line templates/index.qtpl:20
	qw422016 := qt422016.AcquireWriter(qq422016)
//line templates/index.qtpl:20
	StreamIndex(qw422016, page)
//line templates/index.qtpl:20
	qt422016.ReleaseWriter(qw422016)
//line templates/index.qtpl:20
}

//line templates/index.qtpl:20
func Index(page *models.Page) string {
//line templates/index.qtpl:20
	qb422016 := qt422016.AcquireByteBuffer()
//line templates/index.qtpl:20
	WriteIndex(qb422016, page)
//line templates/index.qtpl:20
	qs422016 := string(qb422016.B)
//line templates/index.qtpl:20
	qt422016.ReleaseByteBuffer(qb422016)
//line templates/index.qtpl:20
	return qs422016
//line templates/index.qtpl:20
}