package appinstall

import "html/template"

// NodeJS-specific install/run tokens and template-display config, kept apart from python.go/ruby.go so each type is easy to find and change independently

const nodejsIconSVG = `<i class="ri-nodejs-line text-2xl text-gray-500 dark:text-gray-400" aria-hidden="true"></i>`

var nodeJSDisplay = kindDisplay{
	Icon:  template.HTML(nodejsIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Node.js", RunCommand: "node", RequiredExtension: ".js",
	RequirementsLabel:   "Run NPM install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run npm install using the package.json file, then launch the application. If the application is already built, you can skip this option.",
}

var NodeJS = Kind{
	AppType: "nodejs", DisplayAppType: "NodeJS", PyOrNode: "NODE", Title: "Install Node.js application",
	InstallToken: "npm install", RunToken: "node", DefaultStartupFile: "index.js",
}
