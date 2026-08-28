package appinstall

import "html/template"

// Ruby-specific install/run tokens and template-display config, kept
// apart from nodejs.go/python.go so this type's specifics are easy to find
// and change independently - everything else in this package (routes,
// validation, the shared run-command algorithm) is identical regardless
// of which Kind it's handed.

const rubyIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#CC342D" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-diamond"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 8l4 -5h10l4 5l-11 13z" /><path d="M3 8h18" /><path d="M9 3l2 5l-2.5 10.5" /><path d="M15 3l-2 5l2.5 10.5" /></svg>`

var rubyDisplay = kindDisplay{
	Icon:  template.HTML(rubyIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Ruby", RunCommand: "ruby", RequiredExtension: ".rb",
	RequirementsLabel:   "Run Bundle install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run bundle install using the Gemfile, then launch the application. If the application is already built, you can skip this option.",
}

var Ruby = Kind{
	AppType: "ruby", DisplayAppType: "Ruby", PyOrNode: "RUBY", Title: "Install Ruby Application",
	InstallToken: "bundle install", RunToken: "ruby", DefaultStartupFile: "app.rb",
}
