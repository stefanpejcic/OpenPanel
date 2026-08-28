package appinstall

import "html/template"

// NodeJS-specific install/run tokens and template-display config, kept
// apart from python.go/ruby.go so this type's specifics are easy to find
// and change independently - everything else in this package (routes,
// validation, the shared run-command algorithm) is identical regardless
// of which Kind it's handed.

const nodejsIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#339933" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-brand-nodejs"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 9v8.044a2 2 0 0 1 -2.996 1.734l-1.568 -.9a3 3 0 0 1 -1.436 -2.561v-6.635a3 3 0 0 1 1.436 -2.56l6 -3.667a3 3 0 0 1 3.128 0l6 3.667a3 3 0 0 1 1.436 2.561v6.634a3 3 0 0 1 -1.436 2.56l-6 3.667a3 3 0 0 1 -3.128 0" /><path d="M17 9h-3.5a1.5 1.5 0 0 0 0 3h2a1.5 1.5 0 0 1 0 3h-3.5" /></svg>`

var nodeJSDisplay = kindDisplay{
	Icon:  template.HTML(nodejsIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "NodeJS", RunCommand: "node", RequiredExtension: ".js",
	RequirementsLabel:   "Run NPM install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run npm install using the package.json file, then launch the application. If the application is already built, you can skip this option.",
}

var NodeJS = Kind{
	AppType: "nodejs", DisplayAppType: "NodeJS", PyOrNode: "NODE", Title: "Install NodeJS Application",
	InstallToken: "npm install", RunToken: "node", DefaultStartupFile: "index.js",
}
