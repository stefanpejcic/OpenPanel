package appinstall

import "html/template"

// Python-specific install/run tokens and template-display config, kept apart from nodejs.go/ruby.go so each type is easy to find and change independently

const pythonIconSVG = `<i class="ri-code-s-slash-line text-2xl text-gray-500 dark:text-gray-400" aria-hidden="true"></i>`

var pythonDisplay = kindDisplay{
	Icon:  template.HTML(pythonIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Python", RunCommand: "python", RequiredExtension: ".py",
	RequirementsLabel:   "Run PIP install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run pip install using the requirements.txt file, then launch the application. If the application is already built, you can skip this option.",
}

var Python = Kind{
	AppType: "python", DisplayAppType: "Python", PyOrNode: "PY", Title: "Install Python application",
	InstallToken: "pip install -r requirements.txt", RunToken: "python", DefaultStartupFile: "app.py",
}
