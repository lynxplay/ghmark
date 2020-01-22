package pgk

import (
	"bytes"
	"errors"
	"github.com/shurcooL/github_flavored_markdown"
	"github.com/shurcooL/github_flavored_markdown/gfmstyle"
	"html/template"
	"io/ioutil"
)

const HtmlTemplateString = `
<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8">
             <meta name="viewport" content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
                         <meta http-equiv="X-UA-Compatible" content="ie=edge">
<style>
  {{.Style}}
</style>
<style>
@media print {
  @page { margin: 0; }
  body { margin: 1.6cm; }
}
</style>
</head>
<body class="markdown-body">
  {{.Markdown}}
</body>
</html>
`

func MakeHTMLCompiler() (*HTMLCompiler, error) {
	markdownTemplate, err := template.New("markdown").Parse(HtmlTemplateString)
	if err != nil {
		return nil, errors.New("could not parse html template: " + err.Error())
	}
	cssFile, err := gfmstyle.Assets.Open("gfm.css")
	if err != nil {
		return nil, errors.New("could not open github flavored css file: " + err.Error())
	}
	cssFileContent, err := ioutil.ReadAll(cssFile)
	if err != nil {
		return nil, errors.New("could not read github flavored css file: " + err.Error())
	}
	if err := cssFile.Close(); err != nil {
		return nil, errors.New("could not close github flavored css file: " + err.Error())
	}
	return &HTMLCompiler{
		CSS:          template.CSS(cssFileContent),
		HTMLTemplate: markdownTemplate,
	}, nil
}

type HTMLCompiler struct {
	CSS          template.CSS
	HTMLTemplate *template.Template
}

func (compiler *HTMLCompiler) Compile(fileContent []byte) ([]byte, error) {
	pureMarkdown := string(github_flavored_markdown.Markdown(fileContent))
	templateBuffer := bytes.Buffer{}
	if err := compiler.HTMLTemplate.Execute(&templateBuffer, struct {
		Style    template.CSS
		Markdown template.HTML
	}{
		Style:    compiler.CSS,
		Markdown: template.HTML(pureMarkdown),
	}); err != nil {
		return nil, errors.New("could not execute markdown template: " + err.Error())
	}
	return templateBuffer.Bytes(), nil
}
