package appinstall

import "html/template"

// Ruby-specific install/run tokens and template-display config, kept apart from nodejs.go/python.go so each type is easy to find and change independently

const rubyIconSVG = `<i class="ri-vip-diamond-line text-2xl text-gray-500 dark:text-gray-400" aria-hidden="true"></i>`

var rubyDisplay = kindDisplay{
	Icon:  template.HTML(rubyIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Ruby", RunCommand: "ruby", RequiredExtension: ".rb",
	RequirementsLabel:   "Run bundle install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run bundle install using the Gemfile, then launch the application. If the application is already built, you can skip this option.",
}

var Ruby = Kind{
	AppType: "ruby", DisplayAppType: "Ruby", PyOrNode: "RUBY", Title: "Install Ruby application",
	InstallToken: "bundle install", RunToken: "ruby", DefaultStartupFile: "app.rb",
}
