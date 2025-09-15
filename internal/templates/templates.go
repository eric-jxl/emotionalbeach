package templates

import "embed"

//go:embed index.html
var IndexHTML embed.FS

const DirHTMLContent = `
<!DOCTYPE html>
<html>
<head>
    <title>文件列表</title>
</head>
<body>
    <h1>当前目录：</h1>
    <ul>
        {{ range . }}
            <li>{{ . }}</li>
        {{ end }}
    </ul>
</body>
</html>
`

//go:embed assets/*
var AssetHTML embed.FS
