package appinstall

import "html/template"

// Python-specific install/run tokens and template-display config, kept
// apart from nodejs.go/ruby.go so this type's specifics are easy to find
// and change independently - everything else in this package (routes,
// validation, the shared run-command algorithm) is identical regardless
// of which Kind it's handed.

const pythonIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#3776AB" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-brand-python"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 9h-7a2 2 0 0 0 -2 2v4a2 2 0 0 0 2 2h3" /><path d="M12 15h7a2 2 0 0 0 2 -2v-4a2 2 0 0 0 -2 -2h-3" /><path d="M8 9v-4a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v5a2 2 0 0 1 -2 2h-4a2 2 0 0 0 -2 2v5a2 2 0 0 0 2 2h4a2 2 0 0 0 2 -2v-4" /><path d="M11 6l0 .01" /><path fill="#FFA500" d="M13 18l0 .01" /></svg>`

var pythonDisplay = kindDisplay{
	Icon:  template.HTML(pythonIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Python", RunCommand: "python", RequiredExtension: ".py",
	RequirementsLabel:   "Run PIP install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run pip install using the requirements.txt file, then launch the application. If the application is already built, you can skip this option.",
}

var Python = Kind{
	AppType: "python", DisplayAppType: "Python", PyOrNode: "PY", Title: "Install Python Application",
	InstallToken: "pip install -r requirements.txt", RunToken: "python", DefaultStartupFile: "app.py",
}
